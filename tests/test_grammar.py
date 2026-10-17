import unittest
from grammar import compile_query, summarize, validate


class GrammarTests(unittest.TestCase):
    def test_rejects_unknown_operations_and_injection(self):
        for spec in ({'view': 'sql'}, {'groupBy': 'release; DROP TABLE otel_logs'},
                     {'minutes': True}, {'minutes': 500}, {'sql': 'SELECT 1'},
                     {'scenario': "' OR 1=1"}, []):
            with self.subTest(spec=spec), self.assertRaises(ValueError): validate(spec)

    def test_identifiers_use_query_parameters(self):
        sql, params, _ = compile_query({'reservationId': "x' OR 1=1 --"}, 100000)
        self.assertNotIn(params['id'], sql)
        self.assertIn('{id:String}', sql)

    def test_unknown_does_not_become_success(self):
        result = summarize([{'classification': c} for c in ['good', 'unknown', 'pending']])
        self.assertIsNone(result['successRatio'])
        self.assertEqual(result['alarmPreview'], 'unknown')

    def test_empty_population_is_no_data(self):
        self.assertIsNone(summarize([])['successRatio'])

    def test_native_metrics_cannot_claim_reservation_filter(self):
        with self.assertRaises(ValueError): compile_query({'view': 'metrics', 'reservationId': 'r1'}, 1000)

    def test_span_selection_rejects_invalid_or_unrelated_bounds(self):
        good = {'fromMs': 100, 'toMs': 200, 'minMs': 0, 'maxMs': 10}
        for bounds in ({**good, 'fromMs': True}, {**good, 'toMs': 100},
                       {**good, 'maxMs': -1}, {**good, 'minMs': 10},
                       {**good, 'fromMs': "100 OR 1=1"}, {**good, 'extra': 1},
                       {**good, 'toMs': 9007199254740991}):
            with self.subTest(bounds=bounds), self.assertRaises(ValueError):
                validate({'view': 'traces', 'spanFilter': bounds})
        with self.assertRaises(ValueError): validate({'view': 'logs', 'spanFilter': good})
        sql, params, _ = compile_query({'view': 'traces', 'spanFilter': good}, 100000)
        self.assertIn('{fromMs:Int64}', sql)
        self.assertEqual(params['maxMs'], '10')
        self.assertNotIn('toIntervalMinute', sql)  # selection is frozen, no moving window

    def test_burn_uses_ratio_and_target(self):
        result = summarize([{'classification': 'good'}] * 9 + [{'classification': 'bad'}])
        self.assertAlmostEqual(result['burnRate'], 2)


if __name__ == '__main__': unittest.main()
