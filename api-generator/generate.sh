#! /bin/bash

set -e

if [ -d "api-generator" ]; then
  cd api-generator
fi

GENERATOR=openapi-generator-cli.jar

if [ ! -f "$GENERATOR" ]; then
  echo "Please download https://github.com/Interhyp/openapi-generator/blob/RELTEC-11228/bin/openapi-generator-cli.jar using your browser and save as $GENERATOR"
  exit 1
fi

API_MODEL_PACKAGE_NAME=openapi

java -jar $GENERATOR generate \
  -i ../api/openapi-v3-spec.json \
  -o tmp/$API_MODEL_PACKAGE_NAME \
  --package-name $API_MODEL_PACKAGE_NAME \
  --global-property modelTests=false,modelDocs=false,apiTests=false,apiDocs=false,generateClient=false \
  -g go

# -------------------------------------- customization -----------------------------------------
# omit certain fields from yaml representations, which we use internally to save to files in git
# (this information is represented in the directory tree or is part of the commit metadata)
sed -i'' -e 's/yaml:"timeStamp"/yaml:"-"/g' tmp/$API_MODEL_PACKAGE_NAME/generated_models.go
sed -i'' -e 's/yaml:"commitHash"/yaml:"-"/g' tmp/$API_MODEL_PACKAGE_NAME/generated_models.go
sed -i'' -e 's/yaml:"jiraIssue"/yaml:"-"/g' tmp/$API_MODEL_PACKAGE_NAME/generated_models.go
sed -i'' -e 's/yaml:"owner"/yaml:"-"/g' tmp/$API_MODEL_PACKAGE_NAME/generated_models.go
# ------------------------------------ end customization ---------------------------------------

mkdir -p ../api
mv tmp/$API_MODEL_PACKAGE_NAME/generated_models.go ../api/generated_apimodel.go || (rm -rf tmp && exit 1)
rm -rf tmp

gofmt -w ../api/generated_apimodel.go

DOWNSTREAM_API_DIRECTORY=../internal/client

function generate_downstream {
  P_DOWNSTREAM_NAME=$1
  P_SPEC_FILE_NAME=$2

  MODEL_PACKAGE_NAME=${P_DOWNSTREAM_NAME}client
  java -jar $GENERATOR generate \
    -i ${DOWNSTREAM_API_DIRECTORY}/${P_SPEC_FILE_NAME} \
    -o tmp/${MODEL_PACKAGE_NAME} \
    --package-name ${MODEL_PACKAGE_NAME} \
    --global-property modelTests=false,modelDocs=false,apiTests=false,apiDocs=false \
    -g go
  mkdir -p ${DOWNSTREAM_API_DIRECTORY}/${P_DOWNSTREAM_NAME}
  mv tmp/${MODEL_PACKAGE_NAME}/generated_models.go ${DOWNSTREAM_API_DIRECTORY}/${P_DOWNSTREAM_NAME}/generated_model.go || (rm -rf tmp && exit 1)
  mv tmp/${MODEL_PACKAGE_NAME}/generated_client.go ${DOWNSTREAM_API_DIRECTORY}/${P_DOWNSTREAM_NAME}/generated_client.go || (rm -rf tmp && exit 1)
  rm -rf tmp

  gofmt -w ${DOWNSTREAM_API_DIRECTORY}/${P_DOWNSTREAM_NAME}/generated_model.go
  gofmt -w ${DOWNSTREAM_API_DIRECTORY}/${P_DOWNSTREAM_NAME}/generated_client.go
}

# generate_downstream package_name openapi-spec-filename.json
