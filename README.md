# kakisute

> :warning: このREADME.mdはAIによって生成されたものです。

このリポジトリは、各種技術の検証用プロジェクトをまとめたものです。

## ディレクトリ構成

### Go関連プロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [use-gorm-postgres](./use-gorm-postgres/) | GORMとPostgreSQLの連携検証 | 2024-04-07 |
| [connect-go-example](./connect-go-example/) | Connect Protocol (gRPC) のサンプル実装 | 2024-04-08 |
| [go-1.22-net-http-router](./go-1.22-net-http-router/) | Go 1.22のnet/httpパッケージでのルーター実装検証 | 2024-10-01 |
| [go-serve-mux-group](./go-serve-mux-group/) | Go言語のnet/httpのServeMuxグループ化の検証 | 2024-10-09 |
| [go-parquet-test](./go-parquet-test/) | Parquetファイル形式の読み書き検証 | 2025-01-26 |
| [protoc-gen-gorm-handler](./protoc-gen-gorm-handler/) | Protocol BuffersからGORMハンドラーを生成するツールの検証 | 2025-05-10 |
| [go-http-server-mtls](./go-http-server-mtls/) | mTLSによるHTTPサーバー認証の検証 | 2025-05-16 |
| [go-casbin-test](./go-casbin-test/) | Go言語でのCasbin（認可ライブラリ）の検証 | 2025-05-26 |
| [go-k8s-client-go](./go-k8s-client-go/) | Go言語でのKubernetes APIクライアントの検証 | 2025-05-29 |
| [go-msgpack-test](./go-msgpack-test/) | Go言語でのメッセージパックの検証 | 2025-05-31 |
| [go-echo-error-handling](./go-echo-error-handling/) | Go言語とEchoフレームワークでのエラーハンドリングの検証 | 2025-06-03 |
| [go-otel-echo](./go-otel-echo/) | OpenTelemetryとEchoの統合検証 | 2025-06-12 |
| [go-etcd-embed](./go-etcd-embed/) | 組み込みetcdの検証 | 2025-07-26 |
| [go-generate-client-cert](./go-generate-client-cert/) | Go言語でのクライアント証明書生成の検証 | 2025-10-04 |
| [go-echo-websocket](./go-echo-websocket/) | Go言語とEchoフレームワークでのWebSocketの検証 | 2025-11-11 |
| [go-sakuracloud-monitoringsuite](./go-sakuracloud-monitoringsuite/) | さくらのクラウド監視サービスとの連携 | 2025-11-19 |
| [go-atlas-gorm-test](./go-atlas-gorm-test/) | Go言語でのORM（GORM）とマイグレーションツール（Atlas）の検証 | 2025-11-19 |
| [go-echo-session](./go-echo-session/) | Echoフレームワークでのセッション管理 | 2026-01-11 |
| [go-echo-swagger](./go-echo-swagger/) | Echoフレームワーク + Swaggerの検証 | Unknown |
| [go-smtp-test](./go-smtp-test/) | SMTPサーバーの実装検証 | Unknown |
| [go-client-cert-test](./go-client-cert-test/) | クライアント証明書認証の検証 | Unknown |
| [go-slog-otlp-to-sakura-monitoringsuite](./go-slog-otlp-to-sakura-monitoringsuite/) | slogからOTLP経由でのさくらのクラウド監視サービス送信 | Unknown |


### Web関連プロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [cytoscape-js-react-test](./cytoscape-js-react-test/) | Cytoscape.jsとReactの統合検証 | 2025-07-13 |
| [react-reset-form-use-key](./react-reset-form-use-key/) | Reactでkey属性を使用したフォームリセットの検証 | 2025-10-05 |
| [react-form-test](./react-form-test/) | Reactでのフォーム処理検証 | 2025-12-13 |
| [webapi-media-recorder-test](./webapi-media-recorder-test/) | WebAPI MediaRecorderの検証 | Unknown |

### Python/Djangoプロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [django-wsgi-prom-test](./django-wsgi-prom-test/) | DjangoアプリケーションのPrometheusメトリクス収集検証 | 2025-10-13 |

### その他ツール・インフラ関連

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [tf-backend-http](./tf-backend-http/) | Terraform HTTPバックエンドの検証 | 2024-04-09 |
| [sakuracloud-bridge](./sakuracloud-bridge/) | さくらのクラウドでのブリッジネットワーク設定検証 | 2024-08-12 |
| [kubespray-operations](./kubespray-operations/) | Kubesprayを使用したKubernetesクラスター操作検証 | 2024-08-18 |
| [testloggenerator](./testloggenerator/) | テスト用ログ生成ツール | 2025-03-08 |
| [dependabot-test](./dependabot-test/) | Dependabotの検証 | 2025-04-06 |
| [runn-learn](./runn-learn/) | runnの学習用プロジェクト | 2025-07-05 |
| [otelcol-prom-simple-ha](./otelcol-prom-simple-ha/) | OpenTelemetry CollectorとPrometheusの統合検証 | 2025-08-09 |
| [gdnsd-test](./gdnsd-test/) | gdnsdの検証 | Unknown |

### AI関連プロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [ai-engine-tool-calling](./ai-engine-tool-calling/) | AIエンジンのツール呼び出し検証 | 2025-10-04 |

### さくらのクラウド関連プロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [etcd-on-sakuracloud](./etcd-on-sakuracloud/) | さくらのクラウドでのetcd構築検証 | 2024-11-30 |
| [sakuracloud-simplemq](./sakuracloud-simplemq/) | さくらのクラウドでの簡易MQ検証 | 2025-02-08 |

### さくらのレンタルサーバ関連プロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [sakura-rs-access-log-gen-and-upload-s3](./sakura-rs-access-log-gen-and-upload-s3/) | さくらのレンタルサーバのアクセスログ生成とS3アップロードの検証 | 2025-04-27 |

### データベース関連プロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [pgpool-ii-try](./pgpool-ii-try/) | pgpool-IIの検証 | 2024-06-25 |

### OAuth2認証関連プロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [oauth2-github](./oauth2-github/) | GitHub OAuth2認証の検証 | 2024-04-07 |
| [oauth2-discord](./oauth2-discord/) | Discord OAuth2認証の検証 | 2024-05-25 |
| [oauth2-simpleident](./oauth2-simpleident/) | 簡易的なOAuth2認証サーバー実装 | 2024-09-29 |

### Webアプリケーションプロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [nextcloud](./nextcloud/) | Nextcloudの検証環境 | 2024-07-16 |
| [www](./www/) | WordPressの検証環境 | 2025-05-04 |

### Prometheus関連プロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [prometheus-parse-metrics](./prometheus-parse-metrics/) | Prometheusメトリクスの解析検証 | 2025-06-09 |
| [prometheus-federate-proxy-poc](./prometheus-federate-proxy-poc/) | PrometheusフェデレーションプロキシのPoC | 2025-06-15 |
| [prometheus-federate-proxy](./prometheus-federate-proxy/) | Prometheusフェデレーションプロキシの検証 | 2025-06-15 |

### Nginx関連プロジェクト

| ディレクトリ | 説明 | 追加日付 |
|-------------|------|----------|
| [nginx-blue-green](./nginx-blue-green/) | Nginxを使ったブルー/グリーンデプロイメントの検証 | 2024-07-19 |
## 注意事項

各ディレクトリは独立した検証プロジェクトであり、他のディレクトリとは直接的な依存関係はありません。各プロジェクトの詳細については、それぞれのディレクトリ内のREADMEやソースコードをご参照ください。
