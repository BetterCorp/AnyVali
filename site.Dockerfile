FROM python:3.12-alpine AS build

WORKDIR /app

RUN pip install --no-cache-dir Pillow

COPY tools/site/build_site.py tools/site/build_site.py
COPY site site
COPY logo.png logo.png

RUN python tools/site/build_site.py

FROM docker.angie.software/angie:1.11.4-minimal

COPY docker/angie.conf /etc/angie/angie.conf
COPY --from=build /app/dist-site /usr/share/angie/html

EXPOSE 80

CMD ["angie", "-g", "daemon off;"]
