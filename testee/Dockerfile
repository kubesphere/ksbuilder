FROM alpine:3.15.0

RUN apk add curl bash openssl && \
curl https://raw.githubusercontent.com/helm/chartmuseum/main/scripts/get-chartmuseum | bash && \
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

COPY . /charts
WORKDIR /charts
RUN for file in /charts/*;do if test -d $file;then helm package $file;fi;done && \
helm repo index .

CMD ["chartmuseum","--port=8080",'--storage="local"','--storage-local-rootdir="/charts"']
