#!/bin/bash
set -e

declare -r templates="raw/testconfs/templates"
declare -r output="raw/testconfs/output"
declare -r apikey="raw/api_key.yml"
declare -r spruceAPI="spruce merge ${apikey}"

$spruceAPI "${templates}/onereg.yml" \
           "${templates}/onerate.yml"     > "${output}/oneregonerate.yml" 
$spruceAPI "${templates}/onereg.yml" \
           "${templates}/manyrates.yml"   > "${output}/oneregmanyrates.yml" 
$spruceAPI "${templates}/manyregs.yml" \
           "${templates}/onerate.yml"     > "${output}/manyregsonerate.yml" 

ginkgo -noColor -slowSpecThreshold 8 * 
