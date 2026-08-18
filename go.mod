module github.com/insmtx/corekg

go 1.24.0

// replace github.com/cloudwego/eino => github.com/ygpkg/eino v0.5.2

require (
	github.com/Masterminds/semver v1.5.0
	github.com/aliyun/alibaba-cloud-sdk-go v1.63.107
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.9
	github.com/aws/aws-sdk-go-v2/credentials v1.17.62
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.15.15
	github.com/aws/aws-sdk-go-v2/service/s3 v1.48.1
	github.com/blastrain/vitess-sqlparser v0.0.0-20201030050434-a139afbb1aba
	github.com/cloudwego/eino v0.9.1
	github.com/cloudwego/eino-ext/components/document/loader/url v0.0.0-20250811131559-fa5c34766bc3
	github.com/cloudwego/eino-ext/components/model/openai v0.0.0-20250916084527-de8ccb471c00
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dsnet/golib/memfile v1.0.0
	github.com/elastic/go-elasticsearch/v8 v8.6.0
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/gin-gonic/gin v1.10.1
	github.com/go-sql-driver/mysql v1.8.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/markbates/goth v1.82.0
	github.com/mozillazg/go-pinyin v0.20.0
	github.com/mr-tron/base58 v1.2.0
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/redis/go-redis/v9 v9.7.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/satori/go.uuid v1.2.0
	github.com/silenceper/wechat/v2 v2.1.6
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.11.1
	github.com/swaggo/swag v1.16.6
	github.com/vesoft-inc/nebula-go/v3 v3.8.0
	github.com/xen0n/go-workwx v1.7.0
	github.com/xuri/excelize/v2 v2.9.0
	github.com/ygpkg/yg-go v1.23.74
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.41.0
	golang.org/x/net v0.43.0
	golang.org/x/oauth2 v0.30.0
	golang.org/x/sync v0.16.0
	golang.org/x/text v0.28.0
	google.golang.org/api v0.248.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/gorm v1.30.1
	k8s.io/apimachinery v0.32.8
	k8s.io/client-go v0.32.3
)

require (
	github.com/clbanning/mxj v1.8.4 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/gin-contrib/cors v1.7.2 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.90
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mozillazg/go-httpheader v0.2.1 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/swaggo/files v1.0.1 // indirect
	github.com/swaggo/gin-swagger v1.6.0 // indirect
	github.com/tencentcloud/tencentcloud-cls-sdk-go v1.0.11 // indirect
	github.com/tencentyun/cos-go-sdk-v5 v0.7.58 // indirect
	github.com/vesoft-inc/fbthrift v0.0.0-20230214024353-fa2f34755b28 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gorm.io/driver/mysql v1.5.7
)

require (
	github.com/bytedance/gopkg v0.1.3
	github.com/bytedance/sonic v1.15.0
	github.com/cloudwego/eino-ext/components/embedding/gemini v0.0.0-20260711013131-550170e5a9e0
	github.com/cloudwego/eino-ext/components/embedding/ollama v0.0.0-20260711013131-550170e5a9e0
	github.com/cloudwego/eino-ext/components/embedding/openai v0.0.0-20260711013131-550170e5a9e0
	github.com/getkin/kin-openapi v0.118.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.4.0
	github.com/slongfield/pyfmt v0.0.0-20220222012616-ea85ff4c361f // indirect
	github.com/spf13/cast v1.7.1
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394
	golang.org/x/image v0.22.0
	golang.org/x/mod v0.27.0
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/sqlite v1.5.6
)

require (
	cloud.google.com/go v0.116.0 // indirect
	cloud.google.com/go/auth v0.16.5 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.8.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/99designs/keyring v1.2.1 // indirect
	github.com/AthenZ/athenz v1.12.13 // indirect
	github.com/DataDog/zstd v1.5.0 // indirect
	github.com/JohannesKaufmann/dom v0.2.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/alicebob/gopher-json v0.0.0-20230218143504-906a9b012302 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/anthropics/anthropic-sdk-go v1.56.0 // indirect
	github.com/ardielle/ardielle-go v1.5.2 // indirect
	github.com/avast/retry-go v3.0.0+incompatible // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.16.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.29.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.17 // indirect
	github.com/aws/smithy-go v1.22.2 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.4.0 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874 // indirect
	github.com/buger/jsonparser v1.1.2 // indirect
	github.com/bytedance/sonic/loader v0.5.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/cloudwego/eino-ext/components/document/parser/html v0.0.0-20241224063832-9fbcc0e56c28 // indirect
	github.com/cloudwego/gopkg v0.1.4 // indirect
	github.com/cloudwego/netpoll v0.7.0 // indirect
	github.com/cohesion-org/deepseek-go v1.3.4 // indirect
	github.com/corpix/uarand v0.2.0 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/docker v27.3.0+incompatible // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.0.0-20211216131617-bbee439d559c // indirect
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/evanphx/json-patch v0.5.2 // indirect
	github.com/extrame/ole2 v0.0.0-20160812065207-d69429661ad7 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/frankban/quicktest v1.14.6 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-chi/chi/v5 v5.2.2 // indirect
	github.com/go-jose/go-jose/v4 v4.1.1 // indirect
	github.com/go-json-experiment/json v0.0.0-20250211171154-1ae217ad3535 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-pay/crypto v0.0.1 // indirect
	github.com/go-pay/smap v0.0.2 // indirect
	github.com/go-pay/util v0.0.4 // indirect
	github.com/go-pay/xlog v0.0.3 // indirect
	github.com/go-pay/xtime v0.0.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.27.0 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-shiori/dom v0.0.0-20230515143342-73569d674e1c // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gogs/chardet v0.0.0-20211120154057-b7413eaefb8f // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/goph/emperror v0.17.2 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/hamba/avro/v2 v2.26.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/jsonschema v0.14.0 // indirect
	github.com/invopop/yaml v0.1.0 // indirect
	github.com/itlightning/dateparse v0.2.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/meguminnnnnnnnn/go-openai v0.1.2 // indirect
	github.com/microcosm-cc/bluemonday v1.0.27 // indirect
	github.com/minio/crc64nvme v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.6.0 // indirect
	github.com/nikolalohinski/gonja v1.5.3 // indirect
	github.com/nyaruka/phonenumbers v1.0.55 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/ollama/ollama v0.9.6 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pb33f/ordered-map/v2 v2.3.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/peterbourgon/diskv/v3 v3.0.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pingcap/errors v0.11.5-0.20240311024730-e056997136bb // indirect
	github.com/pingcap/failpoint v0.0.0-20240528011301-b51a646c7c86 // indirect
	github.com/pingcap/log v1.1.1-0.20221015072633-39906604fb81 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.4 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/shabbyrobe/xmlwriter v0.0.0-20200208144257-9fca06d00ffa // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smarty/assertions v1.16.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/standard-webhooks/standard-webhooks/libraries v0.0.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xuri/efp v0.0.0-20240408161823-9ad904a10d6d // indirect
	github.com/xuri/nfp v0.0.0-20240318013403-ab9948c2c4a7 // indirect
	github.com/yargevad/filepathx v1.0.0 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.2 // indirect
	golang.org/x/arch v0.20.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/grpc v1.75.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gorm.io/hints v1.1.0 // indirect
	k8s.io/api v0.32.3 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250701173324-9bd5c66d9911 // indirect
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.6.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
	stathat.com/c/consistent v1.0.0 // indirect
)

require (
	codeberg.org/readeck/go-readability/v2 v2.1.2
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/IBM/sarama v1.45.1
	github.com/JohannesKaufmann/html-to-markdown/v2 v2.4.0
	github.com/PuerkitoBio/goquery v1.10.3
	github.com/alicebob/miniredis/v2 v2.34.0
	github.com/apache/pulsar-client-go v0.16.0
	github.com/apache/rocketmq-client-go/v2 v2.1.3-0.20250427084711-67ec50b93040
	github.com/apache/thrift v0.21.0
	github.com/bytedance/mockey v1.3.0
	github.com/chromedp/cdproto v0.0.0-20250630014756-b7288190f53c
	github.com/chromedp/chromedp v0.13.7
	github.com/cloudwego/eino-ext/components/embedding/ark v0.1.2
	github.com/cloudwego/eino-ext/components/model/ark v0.1.68
	github.com/cloudwego/eino-ext/components/model/claude v0.1.22
	github.com/cloudwego/eino-ext/components/model/deepseek v0.1.6
	github.com/cloudwego/eino-ext/components/model/gemini v0.1.33
	github.com/cloudwego/eino-ext/components/model/ollama v0.1.9
	github.com/cloudwego/eino-ext/components/model/qwen v0.1.9
	github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2 v2.0.0-20251011073417-75b93b87b8a9
	github.com/cloudwego/eino-ext/components/tool/mcp v0.0.6
	github.com/cloudwego/eino-ext/libs/acl/openai v0.1.17
	github.com/cloudwego/hertz v0.10.2
	github.com/ctreminiom/go-atlassian v1.6.1
	github.com/dimchansky/utfbom v1.1.1
	github.com/eino-contrib/jsonschema v1.0.3
	github.com/eino-contrib/ollama v0.1.0
	github.com/extrame/xls v0.0.1
	github.com/go-pay/errgroup v0.0.3
	github.com/go-pay/gopay v1.5.115
	github.com/go-resty/resty/v2 v2.16.5
	github.com/gorilla/sessions v1.2.1
	github.com/hertz-contrib/cors v0.1.0
	github.com/hertz-contrib/sse v0.1.0
	github.com/jaytaylor/html2text v0.0.0
	github.com/jinzhu/copier v0.4.0
	github.com/juju/errors v0.0.0-20170703010042-c7d06af17c68
	github.com/mark3labs/mcp-go v0.43.0
	github.com/matoous/go-nanoid v1.5.1
	github.com/mattn/go-shellwords v1.0.12
	github.com/nats-io/nats.go v1.34.1
	github.com/nsqio/go-nsq v1.1.0
	github.com/onsi/gomega v1.35.1
	github.com/pingcap/tidb/pkg/parser v0.0.0-20250417044355-c5882b1f6c58
	github.com/prometheus/client_golang v1.21.1
	github.com/rbretecher/go-postman-collection v0.9.0
	github.com/slack-go/slack v0.17.3
	github.com/smartystreets/goconvey v1.8.1
	github.com/tealeg/xlsx/v3 v3.3.13
	github.com/tidwall/sjson v1.2.5
	github.com/volcengine/ve-tos-golang-sdk/v2 v2.7.17
	github.com/volcengine/volc-sdk-golang v1.0.211
	github.com/volcengine/volcengine-go-sdk v1.2.30
	github.com/wk8/go-ordered-map/v2 v2.1.8
	github.com/yuin/goldmark v1.7.13
	go.uber.org/mock v0.5.1
	golang.org/x/time v0.12.0
	google.golang.org/genai v1.36.0
	gorm.io/datatypes v1.2.7
	gorm.io/gen v0.3.26
	gorm.io/plugin/dbresolver v1.6.2
)

replace github.com/jaytaylor/html2text => github.com/calebax/html2text v0.1.1

replace github.com/apache/thrift => github.com/apache/thrift v0.13.0

replace go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc => go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0
