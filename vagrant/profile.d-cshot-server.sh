#!/bin/bash
alias cshot-tail="journalctl -f -u cshot-server.service"
alias cshot-rebuild="rm /root/go/bin/cshot-server; go install github.com/apokalyptik/cshot/cmd/cshot-server && systemctl restart cshot-server.service"
