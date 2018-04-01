FROM apextoaster/base:master

ADD ./home-dns /app/home-dns

CMD /app/home-dns /app/config.yml
