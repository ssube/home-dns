FROM apextoaster/base

ADD ./bin/home-dns /app/home-dns

CMD /app/home-dns /app/config.yml
