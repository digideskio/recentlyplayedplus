set -e

spruce merge "raw\\testconfs\\templates\\manyrates.yml" "raw\\api_key.yml" > "raw\\testconfs\\output\\manyrates.yml" 
spruce merge "raw\\testconfs\\templates\\manyregs.yml" "raw\\api_key.yml" > "raw\\testconfs\\output\\manyregs.yml" 
spruce merge "raw\\testconfs\\templates\\manyregsmanyrates.yml" "raw\api_key.yml" > "raw\\testconfs\\output\\manyregsmanyrates.yml" 
spruce merge "raw\\testconfs\\templates\\onereg.yml" "raw\\api_key.yml" > "raw\\testconfs\\output\\onereg.yml" 

ginkgo -noColor -slowSpecThreshold 8 * 
