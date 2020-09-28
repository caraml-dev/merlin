echo "=================================="
echo "============small payload========="
echo "=================================="
cat small.cfg | vegeta attack -rate 100 -duration=60s | vegeta report
cat small.cfg | vegeta attack -rate 100 -duration=60s | vegeta report
cat small.cfg | vegeta attack -rate 100 -duration=60s | vegeta report


echo "=================================="
echo "============medium payload========="
echo "=================================="
cat medium.cfg | vegeta attack -rate 100 -duration=60s | vegeta report
cat medium.cfg | vegeta attack -rate 100 -duration=60s | vegeta report
cat medium.cfg | vegeta attack -rate 100 -duration=60s | vegeta report

echo "=================================="
echo "============large payload========="
echo "=================================="
cat large.cfg | vegeta attack -rate 100 -duration=60s | vegeta report
cat large.cfg | vegeta attack -rate 100 -duration=60s | vegeta report
cat large.cfg | vegeta attack -rate 100 -duration=60s | vegeta report