# AU image
FROM registry.fedoraproject.org/fedora:27

RUN dnf update -y && \
    dnf install -y rpm-ostree && \
    dnf clean all -y

VOLUME /usr/share/rpm
VOLUME /etc
VOLUME /ostree/
VOLUME /boot

ENTRYPOINT ["/usr/bin/rpm-ostree"]
