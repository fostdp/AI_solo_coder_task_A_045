const Charts = (function() {

    function drawBarChart(canvas, data, colorFn) {
        const ctx = canvas.getContext('2d');
        const w = canvas.width;
        const h = canvas.height;
        const padding = { top: 10, right: 10, bottom: 36, left: 36 };
        const chartW = w - padding.left - padding.right;
        const chartH = h - padding.top - padding.bottom;

        ctx.clearRect(0, 0, w, h);

        if (!data || data.length === 0) return;

        const maxVal = Math.max(...data.map(d => d.value));
        const barW = chartW / data.length * 0.7;
        const gap = chartW / data.length * 0.3;

        ctx.strokeStyle = '#3a4050';
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(padding.left, padding.top);
        ctx.lineTo(padding.left, padding.top + chartH);
        ctx.lineTo(padding.left + chartW, padding.top + chartH);
        ctx.stroke();

        ctx.fillStyle = '#8b8b7a';
        ctx.font = '10px sans-serif';
        ctx.textAlign = 'right';
        for (let i = 0; i <= 4; i++) {
            const val = Math.round(maxVal * i / 4);
            const y = padding.top + chartH - chartH * i / 4;
            ctx.fillText(val.toString(), padding.left - 4, y + 3);
            ctx.strokeStyle = 'rgba(58, 64, 80, 0.5)';
            ctx.beginPath();
            ctx.moveTo(padding.left, y);
            ctx.lineTo(padding.left + chartW, y);
            ctx.stroke();
        }

        data.forEach((d, i) => {
            const x = padding.left + i * (barW + gap) + gap / 2;
            const barH = chartH * d.value / maxVal;
            const y = padding.top + chartH - barH;

            const color = colorFn ? colorFn(d.label) : '#d4af37';
            const grad = ctx.createLinearGradient(0, y, 0, y + barH);
            grad.addColorStop(0, color);
            grad.addColorStop(1, shadeColor(color, -30));
            ctx.fillStyle = grad;
            ctx.fillRect(x, y, barW, barH);

            ctx.fillStyle = '#c8c8b8';
            ctx.font = '10px sans-serif';
            ctx.textAlign = 'center';
            ctx.save();
            ctx.translate(x + barW / 2, padding.top + chartH + 14);
            ctx.rotate(-Math.PI / 6);
            ctx.fillText(d.label.length > 6 ? d.label.substring(0, 6) : d.label, 0, 0);
            ctx.restore();
        });
    }

    function drawPieChart(canvas, data) {
        const ctx = canvas.getContext('2d');
        const w = canvas.width;
        const h = canvas.height;

        ctx.clearRect(0, 0, w, h);

        if (!data || data.length === 0) return;

        const centerX = w * 0.4;
        const centerY = h / 2;
        const radius = Math.min(w * 0.35, h * 0.38);
        const total = data.reduce((s, d) => s + d.value, 0);

        let startAngle = -Math.PI / 2;
        data.forEach((d) => {
            const sliceAngle = (d.value / total) * Math.PI * 2;
            ctx.beginPath();
            ctx.moveTo(centerX, centerY);
            ctx.arc(centerX, centerY, radius, startAngle, startAngle + sliceAngle);
            ctx.closePath();
            ctx.fillStyle = d.color || '#d4af37';
            ctx.fill();
            ctx.strokeStyle = '#1a1e27';
            ctx.lineWidth = 2;
            ctx.stroke();
            startAngle += sliceAngle;
        });

        ctx.fillStyle = '#1a1e27';
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius * 0.5, 0, Math.PI * 2);
        ctx.fill();

        ctx.fillStyle = '#f4e5b0';
        ctx.font = 'bold 11px sans-serif';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(total.toString(), centerX, centerY - 6);
        ctx.fillStyle = '#8b8b7a';
        ctx.font = '9px sans-serif';
        ctx.fillText('总计', centerX, centerY + 8);

        const legendX = w * 0.78;
        const legendY = centerY - (data.length * 16) / 2;
        ctx.textAlign = 'left';
        ctx.textBaseline = 'alphabetic';
        data.forEach((d, i) => {
            const y = legendY + i * 16;
            ctx.fillStyle = d.color || '#d4af37';
            ctx.fillRect(legendX, y - 8, 10, 10);
            ctx.fillStyle = '#c8c8b8';
            ctx.font = '10px sans-serif';
            const pct = ((d.value / total) * 100).toFixed(1);
            ctx.fillText(`${d.label} ${pct}%`, legendX + 14, y);
        });
    }

    function drawTerrainProfile(canvas, profile) {
        const ctx = canvas.getContext('2d');
        const w = canvas.width;
        const h = canvas.height;
        const padding = { top: 15, right: 15, bottom: 30, left: 45 };
        const chartW = w - padding.left - padding.right;
        const chartH = h - padding.top - padding.bottom;

        ctx.clearRect(0, 0, w, h);

        if (!profile || !profile.points || profile.points.length === 0) {
            ctx.fillStyle = '#8b8b7a';
            ctx.font = '12px sans-serif';
            ctx.textAlign = 'center';
            ctx.fillText('暂无地形剖面数据', w / 2, h / 2);
            return;
        }

        const pts = profile.points;
        const minElev = profile.min_elev;
        const maxElev = profile.max_elev;
        const elevRange = maxElev - minElev || 1;
        const totalDist = pts[pts.length - 1].distance;
        const distRange = totalDist || 1;

        const bgGrad = ctx.createLinearGradient(0, padding.top, 0, padding.top + chartH);
        bgGrad.addColorStop(0, 'rgba(139, 105, 20, 0.2)');
        bgGrad.addColorStop(1, 'rgba(139, 105, 20, 0.02)');
        ctx.fillStyle = bgGrad;
        ctx.fillRect(padding.left, padding.top, chartW, chartH);

        ctx.strokeStyle = '#3a4050';
        ctx.lineWidth = 0.5;
        ctx.setLineDash([3, 3]);
        for (let i = 0; i <= 4; i++) {
            const y = padding.top + chartH * i / 4;
            ctx.beginPath();
            ctx.moveTo(padding.left, y);
            ctx.lineTo(padding.left + chartW, y);
            ctx.stroke();
        }
        for (let i = 0; i <= 5; i++) {
            const x = padding.left + chartW * i / 5;
            ctx.beginPath();
            ctx.moveTo(x, padding.top);
            ctx.lineTo(x, padding.top + chartH);
            ctx.stroke();
        }
        ctx.setLineDash([]);

        ctx.beginPath();
        pts.forEach((p, i) => {
            const x = padding.left + (p.distance / distRange) * chartW;
            const y = padding.top + chartH - ((p.elevation - minElev) / elevRange) * chartH;
            if (i === 0) ctx.moveTo(x, y);
            else ctx.lineTo(x, y);
        });
        ctx.strokeStyle = '#d4af37';
        ctx.lineWidth = 2;
        ctx.stroke();

        ctx.lineTo(padding.left + chartW, padding.top + chartH);
        ctx.lineTo(padding.left, padding.top + chartH);
        ctx.closePath();
        const fillGrad = ctx.createLinearGradient(0, padding.top, 0, padding.top + chartH);
        fillGrad.addColorStop(0, 'rgba(212, 175, 55, 0.4)');
        fillGrad.addColorStop(1, 'rgba(212, 175, 55, 0.05)');
        ctx.fillStyle = fillGrad;
        ctx.fill();

        ctx.strokeStyle = '#3a4050';
        ctx.lineWidth = 1;
        ctx.setLineDash([]);
        ctx.beginPath();
        ctx.moveTo(padding.left, padding.top);
        ctx.lineTo(padding.left, padding.top + chartH);
        ctx.lineTo(padding.left + chartW, padding.top + chartH);
        ctx.stroke();

        ctx.fillStyle = '#8b8b7a';
        ctx.font = '10px sans-serif';
        ctx.textAlign = 'right';
        for (let i = 0; i <= 4; i++) {
            const val = Math.round(minElev + elevRange * (4 - i) / 4);
            const y = padding.top + chartH * i / 4;
            ctx.fillText(val + 'm', padding.left - 4, y + 3);
        }

        ctx.textAlign = 'center';
        for (let i = 0; i <= 5; i++) {
            const val = Math.round(distRange * i / 5);
            const x = padding.left + chartW * i / 5;
            ctx.fillText(val + 'km', x, padding.top + chartH + 14);
        }

        ctx.fillStyle = '#c8c8b8';
        ctx.font = '11px sans-serif';
        ctx.textAlign = 'left';
        ctx.fillText(`最低: ${Math.round(minElev)}m`, padding.left, 12);
        ctx.textAlign = 'center';
        ctx.fillText(`最高: ${Math.round(maxElev)}m`, w / 2, 12);
        ctx.textAlign = 'right';
        ctx.fillText(`平均: ${Math.round(profile.avg_elev)}m`, padding.left + chartW, 12);
    }

    function shadeColor(color, percent) {
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

    return {
        drawBarChart,
        drawPieChart,
        drawTerrainProfile
    };
})();
