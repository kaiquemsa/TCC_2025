import { Component, Input, OnChanges, ViewChild, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NgxEchartsDirective, NgxEchartsModule } from 'ngx-echarts';
import type { ECharts, EChartsOption, SeriesOption } from 'echarts';
import { ChartSpec } from '../../types/chart';
import { ThemeService } from '../../shared/theme/theme.service';
import * as XLSX from 'xlsx';

@Component({
  selector: 'app-chart-message',
  standalone: true,
  imports: [CommonModule, NgxEchartsModule, NgxEchartsDirective],
  templateUrl: './chart-message.component.html',
  styles: [`
    .font-dark { color: #111827; }
  `]
})
export class ChartMessageComponent implements OnChanges {
  @Input() spec!: ChartSpec;
  @Input() timestamp: Date = new Date();
  option: EChartsOption = {};
  private chartInstance: ECharts | null = null;
  private theme = inject(ThemeService);

  onChartInit(chart: ECharts) {
    this.chartInstance = chart;
  }

  ngOnChanges() { this.build(); }

  private build() {
    const dark = this.theme.theme() === 'dark';
    const text = dark ? '#e5e7eb' : '#111827';
    const grid = dark ? '#334155' : '#e5e7eb';
    const colors = this.spec.colors ?? (dark
      ? ['#60a5fa','#f472b6','#34d399','#fbbf24']
      : ['#2563eb','#db2777','#059669','#d97706']);

    let series: SeriesOption[];

    switch (this.spec.kind) {
      case 'bar':
        series = this.spec.series.map(s => ({
          type: 'bar',
          name: s.name,
          data: s.values,
          stack: this.spec.stacked ? 'stack' : undefined,
        })) as SeriesOption[];
        break;

      case 'line':
        series = this.spec.series.map(s => ({
          type: 'line',
          name: s.name,
          data: s.values,
          smooth: true,
          areaStyle: {}, // remova se não quiser área
        })) as SeriesOption[];
        break;

      case 'pie': {
        // Para pie, usamos a primeira série (ou combine conforme sua regra)
        const first = this.spec.series[0] ?? { name: '', values: [] };
        const pieData = this.spec.x.map((label, i) => ({
          name: label,
          value: first.values[i] ?? 0,
        }));
        series = [{
          type: 'pie',
          name: first.name || this.spec.title,
          radius: '65%',
          data: pieData,
        }] as SeriesOption[];
        break;
      }
    }

    this.option = {
      backgroundColor: 'transparent',
      color: colors,
      textStyle: { color: text },
      tooltip: { trigger: this.spec.kind === 'pie' ? 'item' : 'axis' },
      legend: { textStyle: { color: text } },
      grid: { left: 40, right: 20, top: 40, bottom: 40, containLabel: true },
      xAxis: this.spec.kind === 'pie' ? undefined : {
        type: 'category',
        data: this.spec.x,
        axisLabel: { color: text },
        axisLine: { lineStyle: { color: text } },
      },
      yAxis: this.spec.kind === 'pie' ? undefined : {
        type: 'value',
        name: this.spec.yLabel,
        axisLabel: { color: text },
        splitLine: { lineStyle: { color: grid } },
      },
      series,
    };
  }


  exportAsImage() {
    if (!this.chartInstance) return;
    const dataUrl = this.chartInstance.getDataURL({ type: 'png', backgroundColor: '#fff' });
    const link = document.createElement('a');
    link.href = dataUrl;
    link.download = `${this.spec.title || 'grafico'}.png`;
    link.click();
  }

  exportAsXlsx() {
    if (!this.spec?.series?.length) return;

    const data = this.spec.x.map((label, i) => {
      const row: any = { Label: label };
      this.spec.series.forEach(s => row[s.name] = s.values[i] ?? null);
      return row;
    });

    const ws = XLSX.utils.json_to_sheet(data);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, 'Gráfico');
    XLSX.writeFile(wb, `${this.spec.title || 'grafico'}.xlsx`);
  }

}
