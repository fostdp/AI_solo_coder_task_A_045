const App = (function() {
    let map;
    let canvas;
    let canvasCtx;
    let battlefieldLayer;
    let roadLayer;
    let riverLayer;
    let regionLayer;
    let highProbLayer;

    let filteredBattlefields = [];
    let hoveredBattlefield = null;
    let selectedBattlefield = null;

    let state = {
        showBattlefields: true,
        showRoads: true,
        showRivers: true,
        showRegions: false,
        showHighProb: false,
        filterEra: '',
        filterTerrain: '',
        minTroops: 0,
        useOffscreen: true,
        backgroundType: 'target_group',
        bootstrapRuns: 100
    };

    async function init() {
        initMap();
        initCanvas();
        bindEvents();
        initPerfInfo();
        await AppData.loadAll();
        renderEraLegend();
        renderTroopsLegend();
        drawRoads();
        drawRivers();
        applyFilters();
        renderStats();
        loadFactorAnalysis();
        requestAnimationFrame(renderCanvas);
    }

    function initPerfInfo() {
        const perfStatus = document.getElementById('perf-status');
        if (Charts.isOffscreenCanvasSupported()) {
            perfStatus.innerHTML = '渲染模式: <span class="supported">OffscreenCanvas (硬件加速)</span>';
        } else {
            perfStatus.innerHTML = '渲染模式: <span class="unsupported">标准Canvas (兼容模式)</span>';
            document.getElementById('toggle-offscreen').checked = false;
            state.useOffscreen = false;
        }
    }

    function initMap() {
        map = L.map('map', {
            center: [35, 104],
            zoom: 4,
            minZoom: 3,
            maxZoom: 10,
            zoomControl: true,
            attributionControl: false
        });

        L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
            maxZoom: 19
        }).addTo(map);

        map.on('moveend', () => requestAnimationFrame(renderCanvas));
        map.on('zoomend', () => requestAnimationFrame(renderCanvas));
        map.on('move', () => requestAnimationFrame(renderCanvas));

        roadLayer = L.layerGroup().addTo(map);
        riverLayer = L.layerGroup().addTo(map);
        regionLayer = L.layerGroup();
        highProbLayer = L.layerGroup();
    }

    function initCanvas() {
        canvas = document.getElementById('battlefield-canvas');
        canvasCtx = canvas.getContext('2d');
        resizeCanvas();
        window.addEventListener('resize', resizeCanvas);

        canvas.addEventListener('mousemove', handleMouseMove);
        canvas.addEventListener('click', handleClick);
        canvas.addEventListener('mouseleave', () => {
            hoveredBattlefield = null;
            canvas.style.cursor = 'default';
        });
    }

    function resizeCanvas() {
        const mapContainer = document.getElementById('map');
        canvas.width = mapContainer.clientWidth;
        canvas.height = mapContainer.clientHeight;
        requestAnimationFrame(renderCanvas);
    }

    function bindEvents() {
        document.getElementById('toggle-battlefields').addEventListener('change', (e) => {
            state.showBattlefields = e.target.checked;
            requestAnimationFrame(renderCanvas);
        });
        document.getElementById('toggle-roads').addEventListener('change', (e) => {
            state.showRoads = e.target.checked;
            roadLayer[e.target.checked ? 'addTo' : 'remove'](map);
        });
        document.getElementById('toggle-rivers').addEventListener('change', (e) => {
            state.showRivers = e.target.checked;
            riverLayer[e.target.checked ? 'addTo' : 'remove'](map);
        });
        document.getElementById('toggle-regions').addEventListener('change', (e) => {
            state.showRegions = e.target.checked;
            regionLayer[e.target.checked ? 'addTo' : 'remove'](map);
            if (e.target.checked && AppData.getRegions().length === 0) {
                analyzeRegions();
            }
        });
        document.getElementById('toggle-highprob').addEventListener('change', (e) => {
            state.showHighProb = e.target.checked;
            highProbLayer[e.target.checked ? 'addTo' : 'remove'](map);
            if (e.target.checked && AppData.getHighProbAreas().length === 0) {
                analyzeHighProb();
            }
        });

        document.getElementById('filter-era').addEventListener('change', (e) => {
            state.filterEra = e.target.value;
            applyFilters();
        });
        document.getElementById('filter-terrain').addEventListener('change', (e) => {
            state.filterTerrain = e.target.value;
            applyFilters();
        });
        document.getElementById('filter-troops').addEventListener('input', (e) => {
            state.minTroops = parseInt(e.target.value) || 0;
            document.getElementById('troops-value').textContent = formatTroops(state.minTroops);
            applyFilters();
        });

        document.getElementById('panel-close').addEventListener('click', closePanel);
        document.getElementById('btn-analyze').addEventListener('click', analyzeRegions);
        document.getElementById('btn-highprob').addEventListener('click', analyzeHighProb);

        document.getElementById('background-type').addEventListener('change', (e) => {
            state.backgroundType = e.target.value;
            loadFactorAnalysis();
        });
        document.getElementById('bootstrap-runs').addEventListener('change', (e) => {
            state.bootstrapRuns = parseInt(e.target.value) || 100;
            loadFactorAnalysis();
        });
        document.getElementById('toggle-offscreen').addEventListener('change', (e) => {
            state.useOffscreen = e.target.checked;
            Charts.clearOffscreenCache();
            if (selectedBattlefield) {
                showDetailPanel(selectedBattlefield);
            }
        });
    }

    function applyFilters() {
        const all = AppData.getBattlefields();
        filteredBattlefields = all.filter(bf => {
            if (state.filterEra && bf.era !== state.filterEra) return false;
            if (state.filterTerrain && bf.terrain_type !== state.filterTerrain) return false;
            if (bf.total_troops < state.minTroops) return false;
            return true;
        });
        requestAnimationFrame(renderCanvas);
    }

    function drawRoads() {
        roadLayer.clearLayers();
        AppData.getRoads().forEach(road => {
            const latlngs = road.coords.map(c => [c[1], c[0]]);
            const colors = { '驿道': '#d4af37', '栈道': '#e67e22', '漕运': '#3498db', '官道': '#c0392b', '古道': '#95a5a6' };
            L.polyline(latlngs, {
                color: colors[road.road_type] || '#888',
                weight: road.importance,
                opacity: 0.7,
                dashArray: road.road_type === '古道' ? '5, 5' : null
            }).bindPopup(road.road_name).addTo(roadLayer);
        });
    }

    function drawRivers() {
        riverLayer.clearLayers();
        AppData.getRivers().forEach(river => {
            const latlngs = river.coords.map(c => [c[1], c[0]]);
            const color = river.river_type === '湖泊' ? '#1abc9c' : (river.river_type === '运河' ? '#3498db' : '#2980b9');
            const weight = river.river_type === '湖泊' ? 6 : (river.river_type === '运河' ? 3 : 4);
            L.polyline(latlngs, { color, weight, opacity: 0.7 }).bindPopup(river.river_name).addTo(riverLayer);
        });
    }

    function renderCanvas() {
        if (!canvasCtx || !map) return;
        canvasCtx.clearRect(0, 0, canvas.width, canvas.height);

        if (!state.showBattlefields) return;

        const bounds = map.getBounds();
        filteredBattlefields.forEach(bf => {
            if (!bounds.contains([bf.lat, bf.lng])) return;
            drawTriangle(bf, false);
        });

        if (hoveredBattlefield) {
            drawTriangle(hoveredBattlefield, true);
        }
    }

    function drawTriangle(bf, highlighted) {
        const point = map.latLngToContainerPoint([bf.lat, bf.lng]);
        if (!point) return;

        const size = AppData.getTroopSize(bf.total_troops) * (highlighted ? 1.3 : 1);
        const color = AppData.getEraColor(bf.era);

        canvasCtx.save();
        canvasCtx.translate(point.x, point.y);

        canvasCtx.beginPath();
        canvasCtx.moveTo(0, -size);
        canvasCtx.lineTo(size * 0.866, size * 0.5);
        canvasCtx.lineTo(-size * 0.866, size * 0.5);
        canvasCtx.closePath();

        const grad = canvasCtx.createLinearGradient(0, -size, 0, size * 0.5);
        grad.addColorStop(0, lightenColor(color, 30));
        grad.addColorStop(1, color);
        canvasCtx.fillStyle = grad;
        canvasCtx.fill();

        canvasCtx.strokeStyle = highlighted ? '#f4e5b0' : 'rgba(0,0,0,0.5)';
        canvasCtx.lineWidth = highlighted ? 2 : 1;
        canvasCtx.stroke();

        if (highlighted) {
            canvasCtx.shadowColor = color;
            canvasCtx.shadowBlur = 15;
            canvasCtx.fillStyle = 'rgba(255,255,255,0.3)';
            canvasCtx.fill();
        }

        canvasCtx.restore();
    }

    function handleMouseMove(e) {
        if (!state.showBattlefields) return;
        const rect = canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        hoveredBattlefield = null;
        const bounds = map.getBounds();

        for (let i = filteredBattlefields.length - 1; i >= 0; i--) {
            const bf = filteredBattlefields[i];
            if (!bounds.contains([bf.lat, bf.lng])) continue;
            const pt = map.latLngToContainerPoint([bf.lat, bf.lng]);
            const size = AppData.getTroopSize(bf.total_troops);
            const dx = x - pt.x;
            const dy = y - pt.y;
            if (Math.abs(dx) < size && Math.abs(dy) < size) {
                hoveredBattlefield = bf;
                canvas.style.cursor = 'pointer';
                break;
            }
        }

        if (!hoveredBattlefield) {
            canvas.style.cursor = 'default';
        }
        requestAnimationFrame(renderCanvas);
    }

    function handleClick(e) {
        if (hoveredBattlefield) {
            selectedBattlefield = hoveredBattlefield;
            showDetailPanel(hoveredBattlefield);
        }
    }

    async function showDetailPanel(bf) {
        document.getElementById('panel-title').textContent = bf.battle_name;
        document.getElementById('detail-name').textContent = bf.battle_name;
        document.getElementById('detail-era').textContent = bf.era;
        document.getElementById('detail-dynasty').textContent = bf.dynasty;
        document.getElementById('detail-belligerents').textContent = `${bf.belligerent_a} vs ${bf.belligerent_b}`;
        document.getElementById('detail-troops').textContent = `${formatTroops(bf.troop_a)} : ${formatTroops(bf.troop_b)}（总计${formatTroops(bf.total_troops)}）`;
        document.getElementById('detail-terrain').textContent = bf.terrain_type;
        document.getElementById('detail-elevation').textContent = `${Math.round(bf.elevation)}m`;
        document.getElementById('detail-result').textContent = bf.result;

        document.getElementById('battlefield-panel').classList.remove('hidden');

        try {
            const profile = await fetchWithFallback(
                `/api/terrain_profile?start_lng=${bf.lng - 1}&start_lat=${bf.lat}&end_lng=${bf.lng + 1}&end_lat=${bf.lat}&num_points=50`,
                () => generateMockProfile(bf)
            );
            Charts.drawTerrainProfile(document.getElementById('profile-canvas'), profile, { useOffscreen: state.useOffscreen });
        } catch (e) {
            Charts.drawTerrainProfile(document.getElementById('profile-canvas'), generateMockProfile(bf), { useOffscreen: state.useOffscreen });
        }

        try {
            const acc = await fetchWithFallback(
                `/api/accessibility/${bf.id}`,
                () => generateMockAccessibility(bf)
            );
            renderAccessibility(acc);
        } catch (e) {
            renderAccessibility(generateMockAccessibility(bf));
        }
    }

    function renderAccessibility(acc) {
        const html = `
            <div class="access-score">
                <span class="access-label">交通可达性综合评分</span>
                <span class="access-value">${(acc.accessibility_score * 100).toFixed(1)}</span>
            </div>
            <div class="access-item">
                <span class="access-label">最近道路距离</span>
                <span class="access-value">${acc.nearest_road_dist.toFixed(2)} km</span>
            </div>
            <div class="access-item">
                <span class="access-label">最近道路名称</span>
                <span class="access-value">${acc.nearest_road_name || '未知'}</span>
            </div>
            <div class="access-item">
                <span class="access-label">最近水系距离</span>
                <span class="access-value">${acc.nearest_river_dist.toFixed(2)} km</span>
            </div>
            <div class="access-item">
                <span class="access-label">最近水系名称</span>
                <span class="access-value">${acc.nearest_river_name || '未知'}</span>
            </div>
            <div class="access-item">
                <span class="access-label">10km内道路数</span>
                <span class="access-value">${acc.road_count_in_10km} 条</span>
            </div>
            <div class="access-item">
                <span class="access-label">10km内水系数</span>
                <span class="access-value">${acc.river_count_in_10km} 条</span>
            </div>
        `;
        document.getElementById('accessibility-info').innerHTML = html;
    }

    function closePanel() {
        document.getElementById('battlefield-panel').classList.add('hidden');
        selectedBattlefield = null;
    }

    async function loadFactorAnalysis() {
        try {
            const bg = state.backgroundType;
            const bs = state.bootstrapRuns;
            const data = await fetchWithFallback(
                `/api/site_selection_factors?background=${bg}&bootstrap=${bs}`,
                () => generateMockEnhancedFactors(bg, bs)
            );
            let factors, modelMetrics;
            if (data.factors && data.model_metrics) {
                factors = data.factors;
                modelMetrics = data.model_metrics;
            } else {
                factors = data;
                modelMetrics = null;
            }
            AppData.setFactors(factors);
            if (modelMetrics) {
                renderModelMetrics(modelMetrics);
                document.getElementById('model-metrics').classList.remove('hidden');
            }
            renderFactorAnalysis(factors);
        } catch (e) {
            const fallback = generateMockEnhancedFactors(state.backgroundType, state.bootstrapRuns);
            AppData.setFactors(fallback.factors);
            renderModelMetrics(fallback.model_metrics);
            document.getElementById('model-metrics').classList.remove('hidden');
            renderFactorAnalysis(fallback.factors);
        }
    }

    function renderModelMetrics(metrics) {
        document.getElementById('metric-auc').textContent = (metrics.auc || 0).toFixed(3);
        document.getElementById('metric-accuracy').textContent = ((metrics.accuracy || 0) * 100).toFixed(1) + '%';
        document.getElementById('metric-f1').textContent = (metrics.f1 || 0).toFixed(3);
    }

    function generateMockEnhancedFactors(bg, bs) {
        return {
            factors: [
                { factor_name: '地形高程', contribution: 0.38, p_value: 0.012, odds_ratio: 1.004, stability_score: 0.92, ci95_lower: 1.001, ci95_upper: 1.007, std_err: 0.0015 },
                { factor_name: '交通可达性', contribution: 0.35, p_value: 0.023, odds_ratio: 0.945, stability_score: 0.88, ci95_lower: 0.902, ci95_upper: 0.988, std_err: 0.021 },
                { factor_name: '水源距离', contribution: 0.27, p_value: 0.045, odds_ratio: 0.972, stability_score: 0.76, ci95_lower: 0.948, ci95_upper: 0.996, std_err: 0.012 }
            ],
            model_metrics: {
                auc: 0.842,
                accuracy: 0.785,
                precision: 0.762,
                recall: 0.814,
                f1: 0.787,
                background_type: bg,
                bootstrap_runs: bs
            }
        };
    }

    function renderFactorAnalysis(factors) {
        const container = document.getElementById('factor-analysis');
        container.innerHTML = factors.map(f => {
            const stability = f.stability_score || 0;
            const stabilityClass = stability >= 0.85 ? 'stability-high' : (stability >= 0.7 ? 'stability-medium' : 'stability-low');
            const ci = f.ci95_upper && f.ci95_lower ? `
                <div class="ci-bar">
                    <div class="ci-tick" style="left:0"></div>
                    <div class="ci-tick" style="left:100%"></div>
                    <div class="ci-bar-fill" style="left:25%;right:25%"></div>
                </div>
                <div class="factor-meta" style="font-size:10px;margin-top:2px;color:#8b8b7a">
                    <span>95% CI: [${f.ci95_lower.toFixed(3)}, ${f.ci95_upper.toFixed(3)}]</span>
                </div>
            ` : '';
            const stabilityBadge = f.stability_score ? `
                <span class="stability-indicator ${stabilityClass}" title="Bootstrap稳定性: ${(stability * 100).toFixed(0)}%"></span>
            ` : '';
            return `
            <div class="factor-item">
                <div class="factor-name">
                    <span>${stabilityBadge}${f.factor_name}</span>
                    <span style="color:#d4af37">${(f.contribution * 100).toFixed(1)}%</span>
                </div>
                <div class="factor-bar">
                    <div class="factor-bar-fill" style="width:${f.contribution * 100}%"></div>
                </div>
                ${ci}
                <div class="factor-meta">
                    <span>P值: ${f.p_value.toFixed(3)}</span>
                    <span>OR: ${f.odds_ratio.toFixed(3)}</span>
                    ${f.stability_score ? `<span style="color:#${stability >= 0.85 ? '27ae60' : (stability >= 0.7 ? 'f1c40f' : 'e74c3c')}">稳定: ${(stability * 100).toFixed(0)}%</span>` : ''}
                </div>
            </div>
        `}).join('');
    }

    async function analyzeRegions() {
        const btn = document.getElementById('btn-analyze');
        btn.textContent = '分析中...';
        btn.disabled = true;

        try {
            const data = await fetchWithFallback(
                '/api/military_regions?num_regions=8&fuzzy=true',
                () => generateMockRegionsFCM()
            );
            let regions, fcmResult;
            if (data.regions && data.fcm_result) {
                regions = data.regions;
                fcmResult = data.fcm_result;
            } else {
                regions = data;
                fcmResult = null;
            }
            AppData.setRegions(regions);
            if (fcmResult) {
                renderClusterQuality(fcmResult);
                document.getElementById('cluster-uncertainty').classList.remove('hidden');
                drawRegionsWithUncertainty(regions, fcmResult);
            } else {
                drawRegions(regions);
            }
            if (!map.hasLayer(regionLayer)) {
                regionLayer.addTo(map);
            }
            document.getElementById('stat-regions').textContent = regions.length;
        } finally {
            btn.textContent = '运行军事地理分区';
            btn.disabled = false;
        }
    }

    function renderClusterQuality(fcm) {
        document.getElementById('metric-pc').textContent = (fcm.partition_coef || 0).toFixed(3);
        document.getElementById('metric-pe').textContent = (fcm.partition_entropy || 0).toFixed(3);
        const avgUnc = fcm.uncertainties ? fcm.uncertainties.reduce((a, b) => a + b, 0) / fcm.uncertainties.length : 0;
        document.getElementById('metric-avg-uncertainty').textContent = (avgUnc * 100).toFixed(1) + '%';
    }

    function generateMockRegionsFCM() {
        const mockRegions = generateMockRegions();
        const n = mockRegions.length;
        const uncertainties = new Array(n).fill(0).map(() => 0.05 + Math.random() * 0.25);
        return {
            regions: mockRegions.map((r, i) => ({
                ...r,
                avg_membership: 0.75 + Math.random() * 0.2,
                uncertainty: uncertainties[i]
            })),
            fcm_result: {
                partition_coef: 0.68 + Math.random() * 0.25,
                partition_entropy: 0.15 + Math.random() * 0.3,
                uncertainties: uncertainties,
                membership_matrix: new Array(n).fill(0).map(() =>
                    new Array(8).fill(0).map(() => Math.random()).map((v, i, arr) => v / arr.reduce((a, b) => a + b, 0))
                )
            }
        };
    }

    function drawRegionsWithUncertainty(regions, fcm) {
        regionLayer.clearLayers();
        const colors = ['rgba(231, 76, 60, 0.2)', 'rgba(52, 152, 219, 0.2)', 'rgba(46, 204, 113, 0.2)',
                        'rgba(155, 89, 182, 0.2)', 'rgba(241, 196, 15, 0.2)', 'rgba(230, 126, 34, 0.2)',
                        'rgba(26, 188, 156, 0.2)', 'rgba(211, 84, 0, 0.2)'];
        const borderColors = ['rgba(231, 76, 60, 0.8)', 'rgba(52, 152, 219, 0.8)', 'rgba(46, 204, 113, 0.8)',
                              'rgba(155, 89, 182, 0.8)', 'rgba(241, 196, 15, 0.8)', 'rgba(230, 126, 34, 0.8)',
                              'rgba(26, 188, 156, 0.8)', 'rgba(211, 84, 0, 0.8)'];

        regions.forEach((region, idx) => {
            if (!region.coords || !region.coords[0]) return;
            const latlngs = region.coords[0].map(c => [c[1], c[0]]);
            const uncertainty = region.uncertainty || (fcm.uncertainties ? fcm.uncertainties[idx] : 0.1);
            const baseOpacity = Math.max(0.15, 0.45 - uncertainty * 0.8);
            const dashArray = uncertainty > 0.25 ? '8, 4' : (uncertainty > 0.15 ? '4, 4' : null);

            L.polygon(latlngs, {
                color: borderColors[idx % borderColors.length],
                weight: 2,
                fillColor: colors[idx % colors.length],
                fillOpacity: baseOpacity,
                dashArray: dashArray,
                opacity: Math.max(0.4, 0.8 - uncertainty * 0.8)
            }).bindPopup(`
                <strong>${region.region_name}</strong><br>
                编码: ${region.region_code}<br>
                战役数量: ${region.battle_count}<br>
                战场密度: ${region.avg_density.toFixed(2)}<br>
                主导地形: ${region.dominant_terrain}<br>
                <div style="margin-top:4px;padding-top:4px;border-top:1px solid #333">
                    聚类不确定性: <span class="uncertainty-badge ${uncertainty < 0.15 ? 'uncertainty-low' : (uncertainty < 0.25 ? 'uncertainty-medium' : 'uncertainty-high')}">${(uncertainty * 100).toFixed(1)}%</span><br>
                    平均隶属度: ${((region.avg_membership || 0.8) * 100).toFixed(1)}%
                </div>
            `).addTo(regionLayer);

            if (uncertainty > 0.15) {
                const center = latlngs.reduce((acc, c) => [acc[0] + c[0], acc[1] + c[1]], [0, 0]);
                center[0] /= latlngs.length;
                center[1] /= latlngs.length;
                const ringRadius = 12000 + uncertainty * 40000;
                L.circle(center, {
                    radius: ringRadius,
                    color: 'rgba(231, 76, 60, ' + (0.2 + uncertainty * 0.5) + ')',
                    weight: 1,
                    fillColor: 'rgba(231, 76, 60, ' + (0.05 + uncertainty * 0.15) + ')',
                    fillOpacity: 0.3,
                    interactive: false,
                    className: 'uncertainty-layer'
                }).addTo(regionLayer);
            }

            const center = latlngs.reduce((acc, c) => [acc[0] + c[0], acc[1] + c[1]], [0, 0]);
            center[0] /= latlngs.length;
            center[1] /= latlngs.length;
            L.marker(center, {
                icon: L.divIcon({
                    html: `<div style="background:rgba(0,0,0,0.7);color:#f4e5b0;padding:2px 6px;border-radius:3px;font-size:10px;border:1px solid ${borderColors[idx % borderColors.length]};white-space:nowrap">${region.region_name}</div>`,
                    className: '',
                    iconSize: [80, 18]
                })
            }).addTo(regionLayer);
        });
    }

    function drawRegions(regions) {
        regionLayer.clearLayers();
        const colors = ['rgba(231, 76, 60, 0.2)', 'rgba(52, 152, 219, 0.2)', 'rgba(46, 204, 113, 0.2)',
                        'rgba(155, 89, 182, 0.2)', 'rgba(241, 196, 15, 0.2)', 'rgba(230, 126, 34, 0.2)',
                        'rgba(26, 188, 156, 0.2)', 'rgba(211, 84, 0, 0.2)'];
        const borderColors = ['rgba(231, 76, 60, 0.8)', 'rgba(52, 152, 219, 0.8)', 'rgba(46, 204, 113, 0.8)',
                              'rgba(155, 89, 182, 0.8)', 'rgba(241, 196, 15, 0.8)', 'rgba(230, 126, 34, 0.8)',
                              'rgba(26, 188, 156, 0.8)', 'rgba(211, 84, 0, 0.8)'];

        regions.forEach((region, idx) => {
            if (!region.coords || !region.coords[0]) return;
            const latlngs = region.coords[0].map(c => [c[1], c[0]]);
            L.polygon(latlngs, {
                color: borderColors[idx % borderColors.length],
                weight: 2,
                fillColor: colors[idx % colors.length],
                fillOpacity: 0.4
            }).bindPopup(`
                <strong>${region.region_name}</strong><br>
                编码: ${region.region_code}<br>
                战役数量: ${region.battle_count}<br>
                战场密度: ${region.avg_density.toFixed(2)}<br>
                主导地形: ${region.dominant_terrain}
            `).addTo(regionLayer);

            const center = latlngs.reduce((acc, c) => [acc[0] + c[0], acc[1] + c[1]], [0, 0]);
            center[0] /= latlngs.length;
            center[1] /= latlngs.length;
            L.marker(center, {
                icon: L.divIcon({
                    html: `<div style="background:rgba(0,0,0,0.7);color:#f4e5b0;padding:2px 6px;border-radius:3px;font-size:10px;border:1px solid ${borderColors[idx % borderColors.length]};white-space:nowrap">${region.region_name}</div>`,
                    className: '',
                    iconSize: [80, 18]
                })
            }).addTo(regionLayer);
        });
    }

    async function analyzeHighProb() {
        const btn = document.getElementById('btn-highprob');
        btn.textContent = '分析中...';
        btn.disabled = true;

        try {
            const areas = await fetchWithFallback(
                '/api/high_prob_areas?cell_size=2.5',
                () => generateMockHighProb()
            );
            AppData.setHighProbAreas(areas);
            drawHighProb(areas);
            if (!map.hasLayer(highProbLayer)) {
                highProbLayer.addTo(map);
            }
            document.getElementById('stat-highprob').textContent = areas.length;
        } finally {
            btn.textContent = '分析高概率区域';
            btn.disabled = false;
        }
    }

    function drawHighProb(areas) {
        highProbLayer.clearLayers();
        areas.forEach(area => {
            if (!area.coords || !area.coords[0]) return;
            const latlngs = area.coords[0].map(c => [c[1], c[0]]);
            const prob = area.probability;
            const r = Math.round(255 * prob);
            const g = Math.round(100 * (1 - prob));
            const b = 0;
            L.polygon(latlngs, {
                color: `rgba(${r}, ${g}, 0, 0.6)`,
                weight: 0.5,
                fillColor: `rgba(${r}, ${g}, 0, ${0.3 + prob * 0.4})`,
                fillOpacity: 0.5
            }).bindPopup(`
                <strong>高概率战场区域</strong><br>
                概率: ${(prob * 100).toFixed(1)}%<br>
                地形因子: ${(area.terrain_factor * 100).toFixed(1)}%<br>
                交通因子: ${(area.road_factor * 100).toFixed(1)}%<br>
                水源因子: ${(area.river_factor * 100).toFixed(1)}%
            `).addTo(highProbLayer);
        });
    }

    function renderStats() {
        const all = AppData.getBattlefields();
        document.getElementById('stat-total').textContent = all.length;

        const eraMap = {};
        const terrainMap = {};
        all.forEach(bf => {
            eraMap[bf.era] = (eraMap[bf.era] || 0) + 1;
            terrainMap[bf.terrain_type] = (terrainMap[bf.terrain_type] || 0) + 1;
        });

        const eraData = Object.entries(eraMap).map(([label, value]) => ({ label, value }));
        Charts.drawBarChart(document.getElementById('chart-era'), eraData, (l) => AppData.getEraColor(l));

        const terrainData = Object.entries(terrainMap).map(([label, value]) => ({
            label, value, color: AppData.getTerrainColor(label)
        }));
        Charts.drawPieChart(document.getElementById('chart-terrain'), terrainData);
    }

    function renderEraLegend() {
        const colors = AppData.getEraColors();
        const html = Object.entries(colors).map(([era, color]) => `
            <div class="legend-item">
                <span class="legend-triangle" style="border-bottom:14px solid ${color}"></span>
                <span>${era}</span>
            </div>
        `).join('');
        document.getElementById('era-legend').innerHTML = html;
    }

    function renderTroopsLegend() {
        const levels = [
            { label: '≥50万', troops: 500000 },
            { label: '30-50万', troops: 300000 },
            { label: '15-30万', troops: 150000 },
            { label: '5-15万', troops: 50000 },
            { label: '1-5万', troops: 10000 },
            { label: '<1万', troops: 1000 }
        ];
        const html = levels.map(l => `
            <div class="legend-item">
                <span class="legend-triangle" style="border-bottom:${AppData.getTroopSize(l.troops)}px solid #d4af37; border-left:${AppData.getTroopSize(l.troops)*0.866}px solid transparent; border-right:${AppData.getTroopSize(l.troops)*0.866}px solid transparent; margin: 0 ${(22-AppData.getTroopSize(l.troops)*0.866)}px"></span>
                <span>${l.label}</span>
            </div>
        `).join('');
        document.getElementById('troops-legend').innerHTML = html;
    }

    async function fetchWithFallback(url, fallbackFn) {
        try {
            const res = await fetch(url);
            if (res.ok) return await res.json();
        } catch (e) { }
        return fallbackFn();
    }

    function generateMockProfile(bf) {
        const pts = [];
        const totalDist = 200;
        for (let i = 0; i < 50; i++) {
            const t = i / 49;
            const dist = t * totalDist;
            const elev = bf.elevation + Math.sin(t * Math.PI * 3) * 200 + Math.cos(t * Math.PI * 5) * 80;
            pts.push({ distance: dist, elevation: Math.max(0, elev) });
        }
        const elevs = pts.map(p => p.elevation);
        return {
            start_lng: bf.lng - 1,
            start_lat: bf.lat,
            end_lng: bf.lng + 1,
            end_lat: bf.lat,
            min_elev: Math.min(...elevs),
            max_elev: Math.max(...elevs),
            avg_elev: elevs.reduce((s, e) => s + e, 0) / elevs.length,
            points: pts
        };
    }

    function generateMockAccessibility(bf) {
        const roads = AppData.getRoads();
        const rivers = AppData.getRivers();
        let nrDist = Infinity, nrName = '', nrCount = 0;
        roads.forEach(r => {
            let md = Infinity;
            r.coords.forEach(c => {
                const d = Math.sqrt((c[0] - bf.lng) ** 2 + (c[1] - bf.lat) ** 2) * 100;
                if (d < md) md = d;
            });
            if (md < nrDist) { nrDist = md; nrName = r.road_name; }
            if (md < 10) nrCount++;
        });
        let nrvDist = Infinity, nrvName = '', nrvCount = 0;
        rivers.forEach(r => {
            let md = Infinity;
            r.coords.forEach(c => {
                const d = Math.sqrt((c[0] - bf.lng) ** 2 + (c[1] - bf.lat) ** 2) * 100;
                if (d < md) md = d;
            });
            if (md < nrvDist) { nrvDist = md; nrvName = r.river_name; }
            if (md < 10) nrvCount++;
        });
        return {
            battlefield_id: bf.id,
            nearest_road_dist: nrDist,
            nearest_road_name: nrName,
            nearest_river_dist: nrvDist,
            nearest_river_name: nrvName,
            road_count_in_10km: nrCount,
            river_count_in_10km: nrvCount,
            accessibility_score: Math.random() * 0.5 + 0.3
        };
    }

    function generateMockRegions() {
        const bfs = AppData.getBattlefields();
        const centroids = [
            { lng: 113, lat: 34 }, { lng: 108, lat: 34 }, { lng: 115, lat: 39 },
            { lng: 119, lat: 31 }, { lng: 104, lat: 30 }, { lng: 112, lat: 31 },
            { lng: 101, lat: 38 }, { lng: 122, lat: 41 }
        ];
        const names = ['中原军事区', '关中军事区', '河北军事区', '江南军事区', '巴蜀军事区', '荆襄军事区', '河西军事区', '辽东军事区'];
        const codes = ['ZY', 'GZ', 'HB', 'JN', 'BS', 'JX', 'HX', 'LD'];
        const terrains = ['平原', '河谷', '平原', '河谷', '山地', '关隘', '山地', '平原'];

        return centroids.map((c, i) => {
            const count = bfs.filter(b => Math.sqrt((b.lng - c.lng) ** 2 + (b.lat - c.lat) ** 2) < 8).length || Math.floor(50 + Math.random() * 100);
            const pts = [];
            const numPts = 20;
            for (let j = 0; j < numPts; j++) {
                const angle = 2 * Math.PI * j / numPts;
                const r = 5 + Math.random() * 2;
                pts.push([c.lng + Math.cos(angle) * r, c.lat + Math.sin(angle) * r * 0.8]);
            }
            pts.push(pts[0]);
            return {
                id: i + 1,
                region_name: names[i],
                region_code: codes[i],
                battle_count: count,
                avg_density: count / 200,
                dominant_terrain: terrains[i],
                coords: [pts]
            };
        });
    }

    function generateMockHighProb() {
        const areas = [];
        const hotspots = [{ lng: 113, lat: 34 }, { lng: 108, lat: 34 }, { lng: 115, lat: 37 }, { lng: 119, lat: 31 }, { lng: 104, lat: 30 }];
        let id = 1;
        hotspots.forEach(hs => {
            for (let i = 0; i < 12; i++) {
                const offLng = (Math.random() - 0.5) * 10;
                const offLat = (Math.random() - 0.5) * 8;
                const prob = 0.55 + Math.random() * 0.4;
                const size = 1.5 + Math.random() * 1;
                const lng = hs.lng + offLng;
                const lat = hs.lat + offLat;
                areas.push({
                    id: id++,
                    probability: prob,
                    terrain_factor: prob * 0.38,
                    road_factor: prob * 0.35,
                    river_factor: prob * 0.27,
                    coords: [[
                        [lng, lat], [lng + size, lat], [lng + size, lat + size], [lng, lat + size], [lng, lat]
                    ]]
                });
            }
        });
        return areas;
    }

    function formatTroops(n) {
        if (n >= 10000) return (n / 10000).toFixed(1) + '万';
        return n.toString();
    }

    function lightenColor(color, percent) {
        const num = parseInt(color.replace('#', ''), 16);
        const amt = Math.round(2.55 * percent);
        const R = (num >> 16) + amt;
        const G = (num >> 8 & 0x00FF) + amt;
        const B = (num & 0x0000FF) + amt;
        return '#' + (
            0x1000000 +
            (R < 255 ? (R < 1 ? 0 : R) : 255) * 0x10000 +
            (G < 255 ? (G < 1 ? 0 : G) : 255) * 0x100 +
            (B < 255 ? (B < 1 ? 0 : B) : 255)
        ).toString(16).slice(1);
    }

    return { init };
})();

window.addEventListener('DOMContentLoaded', App.init);
