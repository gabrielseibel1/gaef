FROM rabbitmq:3-management

COPY rabbitmq.conf /etc/rabbitmq/

ENV RABBITMQ_NODENAME=rabbit@localhost

RUN chown rabbitmq:rabbitmq /etc/rabbitmq/rabbitmq.conf

EXPOSE 15672
EXPOSE 5672

USER rabbitmq:rabbitmq