#! /bin/bash

set -e

if [ -d "api-generator" ]; then
  cd api-generator
fi

GENERATOR_VERSION="6.0.1"
GENERATOR="openapi-generator-cli-$GENERATOR_VERSION.jar"

if [ ! -f "$GENERATOR" ]; then
  curl https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/$GENERATOR_VERSION/$GENERATOR > $GENERATOR
fi

API_MODEL_PACKAGE_NAME=openapi

java -jar $GENERATOR generate \
  -i ../docs/openapi-v3-spec.json \
  -o tmp/$API_MODEL_PACKAGE_NAME \
  --package-name $API_MODEL_PACKAGE_NAME \
  --global-property models,modelTests=false,modelDocs=false \
  -g go

(cat tmp/$API_MODEL_PACKAGE_NAME/model_*.go | perl postprocess.pl $API_MODEL_PACKAGE_NAME > ../api/v1/apimodel.go || (rm -rf tmp && exit 1)); rm -rf tmp

gofmt -w ../api/v1/apimodel.go
