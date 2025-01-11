FROM python:3.13-slim

WORKDIR /app

# Install poetry
RUN pip install poetry

# Copy project files
COPY pyproject.toml poetry.lock ./
COPY apis/ ./apis/
COPY commands/ ./commands/
COPY core/ ./core/
COPY cli.py ./

# Install dependencies
RUN poetry config virtualenvs.create false \
    && poetry install --no-dev --no-interaction --no-ansi

# Set entrypoint
ENTRYPOINT ["python", "cli.py"]