module github.com/Interhyp/metadata-service

go 1.21.4

// exclude actually unused dependencies (mostly of pact-go, which is testing only anyway)
// because our scanner fails to understand they are not in use
exclude (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-gonic/gin v1.7.2
	github.com/gogo/protobuf v1.2.1
	github.com/graph-gophers/graphql-go v1.3.0
	github.com/hasura/go-graphql-client v0.6.3
	github.com/prometheus/client_golang v0.9.1
	github.com/prometheus/client_golang v0.9.3
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.1.3
	golang.org/x/net v0.0.0-20180724234803-3673e40ba225
	golang.org/x/net v0.0.0-20180826012351-8a410e7b638d
	golang.org/x/net v0.0.0-20190108225652-1e06a53dbb7e
	golang.org/x/net v0.0.0-20190213061140-3a22650c66bd
	golang.org/x/net v0.0.0-20190311183353-d8887717615a
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3
	golang.org/x/net v0.0.0-20190501004415-9ce7a6920f09
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
	golang.org/x/net v0.0.0-20190603091049-60506f45cf65
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
	golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/net v0.0.0-20200222125558-5a598a2470a0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/net v0.0.0-20200501053045-e0ff5e5a1de5
	golang.org/x/net v0.0.0-20200506145744-7e3656a0809f
	golang.org/x/net v0.0.0-20200513185701-a91f0712d120
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	golang.org/x/net v0.0.0-20210326060303-6b1517762897
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c
	golang.org/x/text v0.3.0
	golang.org/x/text v0.3.1-0.20180807135948-17ff2d5776d2
	golang.org/x/text v0.3.2
)

require (
	github.com/IBM/sarama v1.42.1
	github.com/Roshick/go-autumn-kafka v0.6.0
	github.com/Roshick/go-autumn-synchronisation v0.2.3
	github.com/StephanHCB/go-autumn-config-api v0.2.1
	github.com/StephanHCB/go-autumn-config-env v0.2.2
	github.com/StephanHCB/go-autumn-logging v0.3.0
	github.com/StephanHCB/go-autumn-logging-zerolog v0.5.0
	github.com/StephanHCB/go-autumn-restclient v0.7.3
	github.com/StephanHCB/go-autumn-restclient-apm v0.1.0
	github.com/StephanHCB/go-autumn-restclient-circuitbreaker v0.4.1
	github.com/StephanHCB/go-autumn-restclient-circuitbreaker-prometheus v0.1.0
	github.com/StephanHCB/go-autumn-restclient-prometheus v0.1.2
	github.com/StephanHCB/go-backend-service-common v0.8.0
	github.com/go-chi/chi/v5 v5.0.11
	github.com/go-git/go-billy/v5 v5.5.0
	github.com/go-git/go-git/v5 v5.11.0
	github.com/go-http-utils/headers v0.0.0-20181008091004-fed159eddc2a
	github.com/lestrrat-go/jwx/v2 v2.0.19
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.18.0
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/zerolog v1.31.0
	github.com/stretchr/testify v1.8.4
	gopkg.in/yaml.v3 v3.0.1
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230828082145-3c4c8a2d2371 // indirect
	github.com/StephanHCB/go-autumn-acorn-registry v0.3.1 // indirect
	github.com/StephanHCB/go-autumn-web-swagger-ui v0.2.3 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eapache/go-resiliency v1.5.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/elastic/go-sysinfo v1.7.1 // indirect
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.4 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.19 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/redis/go-redis/v9 v9.3.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546 // indirect
	github.com/skeema/knownhosts v1.2.1 // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	github.com/tidwall/tinylru v1.2.1 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	go.elastic.co/apm/module/apmchiv5/v2 v2.4.5 // indirect
	go.elastic.co/apm/module/apmhttp/v2 v2.4.5 // indirect
	go.elastic.co/apm/v2 v2.4.5 // indirect
	go.elastic.co/fastjson v1.1.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	howett.net/plist v1.0.0 // indirect
)
