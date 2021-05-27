FROM caddy:builder-alpine AS builder

RUN xcaddy build \
    --with github.com/mholt/caddy-ratelimit \
    --with github.com/lolPants/caddy-requestid 

    # TBI
    # --with github.com/greenpau/caddy-auth-jwt \ 
    # --with github.com/greenpau/caddy-auth-portal \
    # --with github.com/greenpau/caddy-trace 
    
FROM caddy:alpine

COPY --from=builder /usr/bin/caddy /usr/bin/caddy