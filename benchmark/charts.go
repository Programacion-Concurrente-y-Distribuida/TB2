package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ─────────────────────────── Estadísticas ─────────────────────────────────────

type boxStats struct {
	Min, Q1, Median, Q3, Max float64
	Outliers                 []float64
}

func calcBox(data []float64) boxStats {
	if len(data) == 0 { return boxStats{} }
	s := make([]float64, len(data))
	copy(s, data)
	sort.Float64s(s)

	q1 := percentile(s, 25)
	med := percentile(s, 50)
	q3 := percentile(s, 75)
	iqr := q3 - q1
	loF := q1 - 1.5*iqr
	hiF := q3 + 1.5*iqr

	wLo, wHi := s[0], s[len(s)-1]
	for _, v := range s { if v >= loF { wLo = v; break } }
	for i := len(s) - 1; i >= 0; i-- { if s[i] <= hiF { wHi = s[i]; break } }

	var out []float64
	for _, v := range s { if v < loF || v > hiF { out = append(out, v) } }
	return boxStats{Min: wLo, Q1: q1, Median: med, Q3: q3, Max: wHi, Outliers: out}
}

func boxToJS(b boxStats) string {
	outs := make([]string, len(b.Outliers))
	for i, v := range b.Outliers { outs[i] = fmt.Sprintf("%.4f", v) }
	return fmt.Sprintf(`{min:%.4f,q1:%.4f,median:%.4f,q3:%.4f,max:%.4f,outliers:[%s]}`,
		b.Min, b.Q1, b.Median, b.Q3, b.Max, strings.Join(outs, ","))
}

func rawToJS(data []float64) string {
	s := make([]string, len(data))
	for i, v := range data { s[i] = fmt.Sprintf("%.4f", v) }
	return "[" + strings.Join(s, ",") + "]"
}

// ─────────────────────────── Generación HTML ──────────────────────────────────

func generateCharts(rows1 []Exp1Row, rows2 []Exp2Row, tSeqTrim float64) {
	banner("REFINANDO GRÁFICOS (PREMIUM VIEW)")

	type ds struct{ times, cpu, mem []float64 }
	dm := map[string]*ds{"Secuencial": {}, "Concurrente": {}}
	for _, r := range rows1 {
		d := dm[r.Paradigm]
		d.times = append(d.times, r.TimeS)
		d.cpu = append(d.cpu, r.CPUPct)
		d.mem = append(d.mem, r.MemPeakMB)
	}

	seq := dm["Secuencial"]
	conc := dm["Concurrente"]

	// JS Data Construction
	jsData := "const tSeqTrim=" + fmt.Sprintf("%.4f", tSeqTrim) + ";\n" +
		"const bSeqT=" + boxToJS(calcBox(seq.times)) + ", bConcT=" + boxToJS(calcBox(conc.times)) + ";\n" +
		"const bSeqC=" + boxToJS(calcBox(seq.cpu)) + ", bConcC=" + boxToJS(calcBox(conc.cpu)) + ";\n" +
		"const bSeqM=" + boxToJS(calcBox(seq.mem)) + ", bConcM=" + boxToJS(calcBox(conc.mem)) + ";\n\n" +
		"const rSeqT=" + rawToJS(seq.times) + ", rConcT=" + rawToJS(conc.times) + ";\n" +
		"const rSeqC=" + rawToJS(seq.cpu) + ", rConcC=" + rawToJS(conc.cpu) + ";\n" +
		"const rSeqM=" + rawToJS(seq.mem) + ", rConcM=" + rawToJS(conc.mem) + ";\n\n"

	var w2L, w2T, w2S []string
	for _, r := range rows2 {
		w2L = append(w2L, fmt.Sprintf("%d", r.Workers))
		w2T = append(w2T, fmt.Sprintf("%.4f", r.TimeTrimS))
		sp := 0.0; if r.TimeTrimS > 0 { sp = tSeqTrim / r.TimeTrimS }
		w2S = append(w2S, fmt.Sprintf("%.4f", sp))
	}
	jsData += "const w2L=[" + strings.Join(w2L, ",") + "], w2T=[" + strings.Join(w2T, ",") + "], w2S=[" + strings.Join(w2S, ",") + "];\n"

	html := htmlHeader() + htmlBody(jsData)
	_ = os.WriteFile(filepath.Join(resultsDir, "charts.html"), []byte(html), 0644)
	fmt.Printf("  ✓ Gráficos premium generados en results/charts.html\n")
}

func htmlHeader() string {
	return `<!DOCTYPE html>
<html lang="es">
<head>
<meta charset="UTF-8"><title>Análisis de Performance Premium</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
<style>
  :root{--bg:#0b0e14;--card:#151921;--brd:#252a34;--seq:#ff5c5c;--conc:#00f2fe;--acc:#bd93f9;--txt:#f8f8f2;--mut:#6272a4}
  body{background:var(--bg);color:var(--txt);font-family:'Segoe UI',Roboto,sans-serif;padding:30px;line-height:1.6}
  h1{text-align:center;font-size:2.2rem;margin-bottom:10px;background:linear-gradient(to right,#ff5c5c,#00f2fe);-webkit-background-clip:text;-webkit-text-fill-color:transparent}
  .sub{text-align:center;color:var(--mut);margin-bottom:40px;text-transform:uppercase;letter-spacing:2px;font-size:0.8rem}
  .grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(300px,1fr));gap:25px;margin-bottom:40px}
  .card{background:var(--card);border:1px solid var(--brd);border-radius:15px;padding:25px;box-shadow:0 10px 30px rgba(0,0,0,0.3)}
  .card h3{font-size:0.9rem;color:var(--acc);margin-bottom:20px;display:flex;align-items:center;gap:10px}
  .card h3::before{content:'';width:4px;height:15px;background:var(--acc);border-radius:2px}
  canvas{width:100%!important;max-height:350px}
  .legend{display:flex;justify-content:center;gap:30px;margin-top:20px}
  .leg-item{display:flex;align-items:center;gap:8px;font-size:0.9rem}
  .dot{width:12px;height:12px;border-radius:3px}
  @media(max-width:800px){.grid{grid-template-columns:1fr}}
</style>
</head>`
}

func htmlBody(jsData string) string {
	return `<body>
<h1>REPORTES DE ALTA PRECISIÓN</h1>
<p class="sub">Comparativa de Paradigmas: Secuencial vs Concurrente</p>

<div class="grid">
  <div class="card"><h3>TIEMPO DE EJECUCIÓN (S) - DISTRIBUCIÓN REAL</h3><canvas id="c-time"></canvas></div>
  <div class="card"><h3>USO DE CPU (%) - DISTRIBUCIÓN REAL</h3><canvas id="c-cpu"></canvas></div>
  <div class="card"><h3>MEMORIA RAM (MB) - DISTRIBUCIÓN REAL</h3><canvas id="c-mem"></canvas></div>
</div>

<div class="legend">
  <div class="leg-item"><div class="dot" style="background:var(--seq)"></div>Secuencial</div>
  <div class="leg-item"><div class="dot" style="background:var(--conc)"></div>Concurrente</div>
</div>

<div class="grid" style="margin-top:50px">
  <div class="card"><h3>ESCALABILIDAD: TIEMPO VS WORKERS</h3><canvas id="l-time"></canvas></div>
  <div class="card"><h3>ESCALABILIDAD: SPEEDUP REAL VS IDEAL</h3><canvas id="l-speed"></canvas></div>
</div>

<script>
` + jsData + `

const COLORS = {seq:'rgba(255,92,92,1)', conc:'rgba(0,242,254,1)', acc:'rgba(189,147,249,1)'};

function makeExp1Chart(id, bSeq, bConc, rSeq, rConc, label) {
    const ctx = document.getElementById(id).getContext('2d');
    
    // Generar Jitter para los puntos raw
    const scatterSeq = rSeq.map(v => ({x: 0.85 + (Math.random()-0.5)*0.1, y: v}));
    const scatterConc = rConc.map(v => ({x: 2.15 + (Math.random()-0.5)*0.1, y: v}));

    new Chart(ctx, {
        type: 'scatter',
        data: {
            datasets: [
                {label: 'Secuencial Raw', data: scatterSeq, backgroundColor: 'rgba(255,92,92,0.2)', pointRadius: 3},
                {label: 'Concurrente Raw', data: scatterConc, backgroundColor: 'rgba(0,242,254,0.2)', pointRadius: 3},
                {label: 'Box', data: [], boxData: [bSeq, bConc]}
            ]
        },
        plugins: [{
            id: 'boxPlugin',
            afterDatasetsDraw(chart) {
                const {ctx, scales: {y}} = chart;
                const xPos = [chart.getDatasetMeta(0).data[0]?.x, chart.getDatasetMeta(1).data[0]?.x];
                if (!xPos[0]) return;

                [bSeq, bConc].forEach((b, i) => {
                    const x = xPos[i];
                    const col = i === 0 ? COLORS.seq : COLORS.conc;
                    const w = 40;
                    
                    ctx.save();
                    ctx.strokeStyle = col; ctx.lineWidth = 2;
                    ctx.strokeRect(x - w/2, y.getPixelForValue(b.q3), w, y.getPixelForValue(b.q1) - y.getPixelForValue(b.q3));
                    ctx.beginPath();
                    ctx.moveTo(x - w/2, y.getPixelForValue(b.median)); ctx.lineTo(x + w/2, y.getPixelForValue(b.median));
                    ctx.lineWidth = 4; ctx.stroke();
                    
                    ctx.lineWidth = 1;
                    [[b.min, b.q1], [b.q3, b.max]].forEach(([y1, y2]) => {
                        ctx.beginPath(); ctx.moveTo(x, y.getPixelForValue(y1)); ctx.lineTo(x, y.getPixelForValue(y2)); ctx.stroke();
                    });
                    ctx.restore();
                });
            }
        }],
        options: {
            scales: {
                x: {display: false, min: 0.5, max: 2.5},
                y: {title: {display: true, text: label, color: '#6272a4'}, grid: {color: '#252a34'}, ticks: {color: '#6272a4'}}
            },
            plugins: {legend: {display: false}}
        }
    });
}

makeExp1Chart('c-time', bSeqT, bConcT, rSeqT, rConcT, 'Segundos');
makeExp1Chart('c-cpu', bSeqC, bConcC, rSeqC, rConcC, 'Porcentaje %');
makeExp1Chart('c-mem', bSeqM, bConcM, rSeqM, rConcM, 'Megabytes MB');

// Charts de Lineas
new Chart(document.getElementById('l-time'), {
    type: 'line',
    data: {
        labels: w2L,
        datasets: [{
            label: 'Tiempo Concurrente', data: w2T, borderColor: COLORS.conc, backgroundColor: 'rgba(0,242,254,0.1)', fill: true, tension: 0.3, pointRadius: 5
        }, {
            label: 'Referencia Secuencial', data: Array(w2L.length).fill(tSeqTrim), borderColor: COLORS.seq, borderDash: [5,5], pointRadius: 0
        }]
    },
    options: {
        scales: {
            x: {grid: {color: '#252a34'}, ticks: {color: '#6272a4'}},
            y: {grid: {color: '#252a34'}, ticks: {color: '#6272a4'}}
        },
        plugins: {legend: {labels: {color: '#f8f8f2'}}}
    }
});

new Chart(document.getElementById('l-speed'), {
    type: 'line',
    data: {
        labels: w2L,
        datasets: [{
            label: 'Speedup Real', data: w2S, borderColor: COLORS.acc, backgroundColor: 'rgba(189,147,249,0.1)', fill: true, tension: 0.3, pointRadius: 5
        }, {
            label: 'Speedup Ideal', data: w2L.map(w => w/w2L[0]), borderColor: '#44475a', borderDash: [5,5], pointRadius: 0
        }]
    },
    options: {
        scales: {
            x: {grid: {color: '#252a34'}, ticks: {color: '#6272a4'}},
            y: {grid: {color: '#252a34'}, ticks: {color: '#6272a4'}}
        },
        plugins: {legend: {labels: {color: '#f8f8f2'}}}
    }
});
</script>
</body></html>`
}
