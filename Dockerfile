FROM     scratch

ADD      azkube      /opt/azkube/azkube
ADD      templates   /opt/azkube/templates

WORKDIR  /opt/azkube
CMD      [ "/opt/azkube/azkube" ]
