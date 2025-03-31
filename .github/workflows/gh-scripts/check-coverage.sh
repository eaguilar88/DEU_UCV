THRESHOLD="5.0"

grep -v -E '(/gen/|/generate/|mock|test|/common/proto/)' coverage.tmp.out > coverage.out
COVERAGE=$(go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+')
echo -e "\nCoverage is $COVERAGE.\n"
if (( $(echo "$COVERAGE >= $THRESHOLD" | bc -l) )); then
  echo -e "Coverage check passed.\n"
else
  echo -e "Error: Coverage below threshold ($COVERAGE < $THRESHOLD).\n"
  exit 1
fi
