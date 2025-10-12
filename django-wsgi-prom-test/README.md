
## 検証手順

```
uv init
uv add django
uv add --group prod gunicorn
uv sync
uv run django-admin startproject prom_test .
```