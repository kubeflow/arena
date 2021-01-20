#!/bin/bash

rm .github/dependabot.yml
cp .github/base_dependabot.yml.tmp .github/dependabot.yml

for directory in $(dirname $(find . -type f -name "*ockerfile*") | sort -u | cut -c2-); do
    yq eval -i ".updates += {\"package-ecosystem\":\"docker\",\"directory\":\"${directory}\",\"schedule\":{\"interval\":\"daily\"},\"open-pull-requests-limit\":10}" .github/dependabot.yml
done

