FROM python:3.12-alpine AS build

WORKDIR /app

COPY tools/docs/requirements.txt tools/docs/requirements.txt
RUN pip install --no-cache-dir -r tools/docs/requirements.txt

COPY mkdocs.yml mkdocs.yml
COPY tools/docs/build_docs.py tools/docs/build_docs.py
COPY README.md README.md
COPY CONTRIBUTING.md CONTRIBUTING.md
COPY LICENSE LICENSE
COPY docs docs
COPY spec spec

RUN python tools/docs/build_docs.py
RUN mkdocs build --strict --site-dir dist-docs

FROM docker.angie.software/angie:1.11.4-minimal

COPY docker/angie.conf /etc/angie/angie.conf
COPY --from=build /app/dist-docs /usr/share/angie/html

EXPOSE 80

CMD ["angie", "-g", "daemon off;"]
