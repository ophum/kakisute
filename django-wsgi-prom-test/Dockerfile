FROM python:3.12

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONBUFFERED=1 \
    UV_LINK_MODE=copy \
    UV_PYTHON_DOWNLOADS=never \
    UV_PROJECT_ENVIRONTMENT=/app/.venv


COPY --from=ghcr.io/astral-sh/uv:0.9.2 /uv /uvx /bin/

COPY . .
RUN uv sync --frozen --no-cache --group prod

