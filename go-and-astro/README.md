# Go and Astro

Astro で生成した HTML を Go の `html/template` として読み込み、Echo から配信する実験です。フロントエンドには React と Tailwind CSS も使用しています。

## 必要な環境

- Go 1.27.4 以上
- Node.js 22.23.4 以上と npm

## 起動

`go-and-astro` ディレクトリで実行します。

```sh
cd frontend
npm ci
npm run build
cd ..
go run .
```

ブラウザで <http://localhost:8080/> を開きます。Go のビルド時に `frontend/dist` を埋め込むため、初回起動時やフロントエンドの変更後は先に `npm run build` を実行してください。

## 仕組み

- Astro は `frontend/dist` に静的 HTML とアセットを生成します。
- `frontend/src/components/Go.astro` は `{{ .Name }}` などの Go テンプレート構文を出力します。Go サーバーは生成された HTML をテンプレートとして解析し、リクエスト時に値を渡して描画します。
- `/` は名前の一覧と Markdown から生成された投稿の一覧を表示し、`/form` は名前の入力画面を表示します。`/posts/:name` は対応する投稿を表示します。
- 名前のデータはサーバーのメモリ内に保持され、再起動すると消えます。

## 開発時の自動再読み込み

[Air](https://github.com/air-verse/air) をインストールしている場合、`go-and-astro` ディレクトリで `air` を実行すると、設定に従って Astro のビルドと Go サーバーの再ビルドが行われます。
