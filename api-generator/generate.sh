#! /bin/bash

set -e

if [ -d "api-generator" ]; then
  cd api-generator
fi

GENERATOR_VERSION=7.7.1_INTERHYP
GENERATOR_NAME=openapi-generator-cli
GENERATOR=$GENERATOR_NAME-$GENERATOR_VERSION.jar

if [ ! -f "$GENERATOR" ]; then
  curl -k https://raw.githubusercontent.com/Interhyp/openapi-generator/refs/heads/new_generator_rebased/bin/$GENERATOR > $GENERATOR
fi

API_MODEL_PACKAGE_NAME=openapi

function generate_apimodel {
  java -jar $GENERATOR generate \
    -i ../api/openapi-v3-spec.yaml \
    -o ../api \
    --package-name $API_MODEL_PACKAGE_NAME \
    --global-property models \
    --additional-properties=enumClassPrefix=true,structPrefix=true \
    -g go-autumrest
}

DOWNSTREAM_API_DIRECTORY=../internal/client

function generate_downstream {
  P_DOWNSTREAM_NAME=$1
  P_SPEC_FILE_NAME=$2
  # use 'tags' from openapi to generate only selected parts of the entire api. Use ':' as separator for multiple values. Convert whitespaces to CamelCased string: 'Abc and Def'->'AbcAndDef'
  P_APIS=$3

  MODEL_PACKAGE_NAME=${P_DOWNSTREAM_NAME}client
  java -jar $GENERATOR generate \
    -i ${DOWNSTREAM_API_DIRECTORY}/${P_SPEC_FILE_NAME} \
    -o ${DOWNSTREAM_API_DIRECTORY}/${P_DOWNSTREAM_NAME} \
    --package-name ${MODEL_PACKAGE_NAME} \
    --global-property supportingFiles,models,apis=${P_APIS} \
    --additional-properties=enumClassPrefix=true,structPrefix=true \
    -g go-autumrest
}

generate_apimodel
generate_downstream bitbucket bitbucket-v8.19.json BuildsAndDeployments:PullRequests:Repository:User

# -------------------------------------- customization -----------------------------------------
# omit certain fields from yaml representations, which we use internally to save to files in git
# (this information is represented in the directory tree or is part of the commit metadata)
for i in ../api/*.go; do
    sed -i'' -e 's/yaml:"timeStamp"/yaml:"-"/g' $i
    sed -i'' -e 's/yaml:"commitHash"/yaml:"-"/g' $i
    sed -i'' -e 's/yaml:"jiraIssue"/yaml:"-"/g' $i
    sed -i'' -e 's/yaml:"owner"/yaml:"-"/g' $i
done

# ------------------------------------ end customization ---------------------------------------

gofmt -w ../api/*.go
