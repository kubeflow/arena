# Delete the serving jobs

If you want to delete serving jobs, the command ``arena serve delete`` can help you.

List all serving jobs:

    $ arena serve list
    NAME                 TYPE        VERSION       DESIRED  AVAILABLE  ADDRESS       PORTS
    fast-style-transfer  Custom      alpha         1        1          172.28.14.93  RESTFUL:31129->5000
    mymnist1             Tensorflow  202101162119  1        0          172.28.3.123  GRPC:8500,RESTFUL:8501

Delete the serving jobs:

    $ arena serve delete fast-style-transfer  mymnist1
