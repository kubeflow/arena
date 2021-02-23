# Get serving job logs

You can use ``arena serve logs`` to get the serving job logs, we will introduce some usages for you.

1\. ``arena serve logs`` can help you to get the serving job logs.

    $ arena serve logs fast-style-transfer
    * Serving Flask app "app" (lazy loading)
    * Environment: production
    WARNING: This is a development server. Do not use it in a production deployment.
    Use a production WSGI server instead.
    * Debug mode: off
    * Running on http://0.0.0.0:5000/ (Press CTRL+C to quit)

2\. If you want to get the last n lines of the serving job logs, you can use ``-t n``(or ``--tail n``), the following command will display the last 5 lines of the serving job.

    $ arena serve logs fast-style-transfer -t 5
    * Environment: production
    WARNING: This is a development server. Do not use it in a production deployment.
    Use a production WSGI server instead.
    * Debug mode: off
    * Running on http://0.0.0.0:5000/ (Press CTRL+C to quit)

3\. If you want to get the logs of target serving job instance, you can use ``-i <INSTANCE_NAME>``(or ``--instance <INSTANCE_NAME>``).

Get the instance of serving job from ``arena serve get`` command.

    $ arena serve get fast-style-transfer
    Name:           fast-style-transfer
    Namespace:      default
    Type:           Custom
    Version:        alpha
    Desired:        1
    Available:      1
    Age:            22m
    Address:        172.28.14.93
    Port:           RESTFUL:31129->5000
    GPUs:           1

    Instances:
    NAME                                                       STATUS   AGE  READY  RESTARTS  GPUs  NODE
    ----                                                       ------   ---  -----  --------  ----  ----
    fast-style-transfer-alpha-custom-serving-856dbcdbcb-sxx2n  Running  22m  1/1    0         1     cn-beijing.192.168.1.112

Get the logs of training job instance by specifying option ``-i``.

    $ arena serve logs fast-style-transfer -i  fast-style-transfer-alpha-custom-serving-856dbcdbcb-sxx2n
    * Serving Flask app "app" (lazy loading)
    * Environment: production
    WARNING: This is a development server. Do not use it in a production deployment.
    Use a production WSGI server instead.
    * Debug mode: off
    * Running on http://0.0.0.0:5000/ (Press CTRL+C to quit)


4\. If you want to get a serving job logs in a duration, for example, ``--since 5m`` represents that getting the serving job logs from five minutes ago to the present.

    $ arena serve logs tf-serving-test --since 5m

5\. If you want to real-time display the serving job logs, ``-f`` is required.

    $ arena serve logs tf-serving-test -f