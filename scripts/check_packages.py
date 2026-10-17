"""Executable CLI storyboard: bind -> expand -> prepare -> evaluate/SQL parity.
Uses normalized fixtures and embedded ClickHouse, not a live HTTP source adapter.
"""
import json
from pathlib import Path
import subprocess
import tempfile
import yaml
from chdb import session

ROOT=Path(__file__).resolve().parents[1]
with tempfile.TemporaryDirectory(prefix='goo-package-') as directory:
    temp=Path(directory)
    binary=temp/'goo'
    subprocess.run(['go','build','-o',str(binary),'./cmd/goo'],cwd=ROOT,check=True)
    common=['-binding',str(ROOT/'examples/packages/http.yaml'),'-inventory',str(ROOT/'examples/packages/inventory.json')]
    def cli(command,*args):
        return json.loads(subprocess.check_output([str(binary),command,*args],cwd=ROOT,text=True))
    expanded=cli('expand',*common)
    assert expanded['plans']['clickhouse']['status']=='requires-prepared-source'
    assert any(d['capability']=='duration-heatmap-and-waterfall' and d['status']=='unsupported' for d in expanded['diagnostics'])
    definition=temp/'measurement.yaml'
    definition.write_text(yaml.safe_dump(expanded['measurement']))
    prepared=cli('prepare-http',*common,'-requests',str(ROOT/'examples/packages/requests.json'),'-evaluation','http-1')
    assert prepared['expansionRevision']==expanded['revision']
    facts=temp/'facts.json';facts.write_text(json.dumps(prepared['facts']))
    ref=cli('evaluate','-definition',str(definition),'-facts',str(facts),'-evaluation','http-1','-group-by','route,method')
    db=session.Session(str(temp/'db'))
    try:
        db.query('CREATE TABLE prepared_http (evaluation_id String, scope_id String, trace_id String, span_id String, route String, method String, classification String) ENGINE=Memory')
        db.query('INSERT INTO prepared_http FORMAT JSONEachRow\n'+'\n'.join(json.dumps(f) for f in prepared['facts']))
        query=expanded['plans']['clickhouse']['query'].replace('{evaluation_id:String}',"'http-1'").replace('{evidence_impaired:Bool}','false')
        rows=[json.loads(line) for line in str(db.query(query)).splitlines() if line]
        # Independent expectations: source fixture contains 200,503,null and 404.
        expected={('/checkout','POST'):(3,1,1,1,'unknown',None),('(unknown route)','GET'):(1,1,0,0,'ok',1.0)}
        for row in rows:
            key=(row['route'],row['method']);want=expected[key]
            got=tuple(int(row[n]) for n in ('requests','successful','failures','unknown'))+(row['success_ratio_state'],row['success_ratio'])
            assert got==want,(got,want)
        assert len(rows)==len(expected)
        for row in ref:
            m=row['measures'];key=(row['dimensions']['route'],row['dimensions']['method']);want=expected[key]
            got=tuple(m[n]['count'] for n in ('requests','successful','failures','unknown'))+(m['success_ratio']['state'],m['success_ratio'].get('ratio'))
            assert got==want,(got,want)
        metric=yaml.safe_load(expanded['plans']['databricks']['metricYaml'])
        assert metric['measures'] and metric['dimensions']
    finally:
        db.close()
    print('PASS: package CLI storyboard, two independent route expectations, reference/ClickHouse parity, Databricks YAML structure')
