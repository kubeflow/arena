# FAQ: Failed To Install Arena on Mac

#### Problem
I need to install arena on my mac, but an error occurred and the error message is:

![error message](./error_message_1.jpg)

#### Solution

1\. Run the following command to disable "Gatekeeper".

```
$ sudo spctl --master-disable
```

2\. Install the arena try again.