FROM rocker/r-base

LABEL io.openshift.s2i.scripts-url="image:///s2i/bin"

RUN apt-get update -qq && apt-get install -y \
      git-core \
      libgit2-dev \
      libsodium-dev \
      libssl-dev \
      libcurl4-gnutls-dev

RUN Rscript -e "install.packages('plumber')"
RUN Rscript -e "install.packages('optparse')"
RUN Rscript -e "install.packages('jsonlite')"
RUN Rscript -e "install.packages('urltools')"
RUN Rscript -e "install.packages('caret')"
RUN Rscript -e "install.packages('devtools')"

RUN mkdir microservice
WORKDIR /microservice

COPY microservice.R /microservice

COPY ./s2i/bin/ /s2i/bin

EXPOSE 5000
