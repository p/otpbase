# https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/
FROM scratch
ADD tmp/otpbase.docker /otpbase.docker
ENV PORT 80
CMD ["/otpbase.docker"]
