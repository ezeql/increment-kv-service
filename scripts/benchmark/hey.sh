#!/bin/bash
hey -c 200 -n 50000 -m POST -T 'application/json' -d '{"key": "12345678-1234-5678-1234-567812345678","value": 1}' http://localhost:3333/increment   
