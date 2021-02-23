# Clean the finished training jobs

You can clean the finished the training jobs only use command ``arena prune --since <DURATION>``.

the following command is an example, it represents that arena will clean up the training jobs whose status are ``FAILED`` or ``COMPLETEDED`` 5 minutes ago.

    $ arena prune --since 5m


If you want to clean all training jobs of all namespaces, you should add option ``--all-namespaces``.

    $ arena prune --all-namespaces