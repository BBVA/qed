#!/usr/bin/env python 

from plotly import tools
import plotly.plotly as py 
import plotly.graph_objs as go
from plotly.offline import download_plotlyjs, init_notebook_mode, plot, iplot
import sys
import json
import math

def main(argv):
    inputfile = sys.argv[1]
    with open(inputfile) as f:
        lines = f.readlines()
        metrics = [json.loads(line) for line in lines] 
        
        minutes = [i for i, x in enumerate(metrics, 1)]

        count_metrics = []
        for m in metrics:
            for e in m:
                if e['name'] == 'qed_hyper_add_total':
                    count_metrics.append(e['metrics'][0])
        #count_metrics = [y['metrics'][0] for y in x if (y['name'] == 'qed_hyper_add_total') for x in metrics]
        counts = [0] + [float(x['value']) for x in count_metrics]
        counts_deltas = [max(0, y - x) for x, y in zip(counts, counts[1:])]
        rates = [x/60 for x in counts_deltas]

        trace_rates = go.Scatter(
            x=minutes, 
            y=rates,
            name='Rate'
        )   

        trace_events = go.Scatter(
            x=minutes,
            y=counts,
            name="Num events"
        )  

        fig = tools.make_subplots(rows=2, cols=1, specs=[[{}],[{}]],
                                subplot_titles=('Rates (events/sec)', 'Num events')
                                )

        fig.append_trace(trace_rates, 1, 1)
        fig.append_trace(trace_events, 2, 1)

        fig.layout['yaxis1'].update(range=[0, 20000])
        
        plot(fig)

if __name__ == "__main__":
    main(sys.argv)