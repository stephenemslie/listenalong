FROM python:3.8.3-slim-buster
WORKDIR /usr/src/app
RUN set -eux; \
    apt-get update; \
    apt-get install -y curl; \
    curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py | python
RUN echo "export PATH=\"$HOME/.poetry/bin:$PATH\"" >> ~/.bashrc
COPY pyproject.toml .
RUN $HOME/.poetry/bin/poetry install
ENTRYPOINT ["./docker-entrypoint.sh"]
CMD ["runserver"]
