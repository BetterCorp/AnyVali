FROM python:3.12-alpine AS build

WORKDIR /app

COPY docs-site/requirements.txt docs-site/requirements.txt
RUN pip install --no-cache-dir -r docs-site/requirements.txt

COPY mkdocs.yml mkdocs.yml
COPY docs-site/build_docs.py docs-site/build_docs.py
COPY README.md README.md
COPY CONTRIBUTING.md CONTRIBUTING.md
COPY LICENSE LICENSE
COPY docs docs
COPY spec spec

RUN python docs-site/build_docs.py
RUN mkdocs build --strict

FROM docker.angie.software/angie:1.11.4-minimal

COPY docker/angie.conf /etc/angie/angie.conf
COPY --from=build /app/site /usr/share/angie/html

EXPOSE 80

CMD ["angie", "-g", "daemon off;"]
