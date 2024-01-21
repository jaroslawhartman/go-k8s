# Intro

A small PoC to manipulate Kubernetes API. In this app:
 * Create a `nginx` Pod
 * Wait the pod to be Running
 * Execute `ls -l` and capture the output in `buffer`  (`strings.Builder`)
 * Scan for a line with `media` and print it

# ToDo

 * File transfer to/from a Pod