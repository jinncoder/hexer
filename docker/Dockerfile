FROM golang:alpine AS runtime

COPY --from=builder /build/dist/temp/hexer-linux_amd64 /usr/local/bin/hexer

RUN chmod +rx /usr/local/bin/hexer && mkdir /app/
RUN apk add --no-cache openssh-keygen

WORKDIR /app
USER root

COPY --chown=hexer <<EOF /app/.hexer.yaml
storage: /app/storage
local: false
ip: 0.0.0.0
port: 4242
database_ip: 0.0.0.0
database_port: 8090
theme: # base from https://ethanschoonover.com/solarized/
  style:
    header_text:
      foreground:
        light: "#859900"
        dark: "#859900"
    status:
      foreground:
        light: "#859900"
        dark: "#859900"
    status_header:
      foreground:
        light: "#586E75"
        dark: "#93A1A1"
    title_style:
      foreground:
        light: "#859900"
        dark: "#859900"
      background:
        light: "#586E75"
        dark: "#93A1A1"
    error_header_text:
      foreground:
        light: "#C24543"
        dark: "#C24543"
    tab_style:
      inactive_highlight_color:
        border_foreground:
          light: "#457A9F"
          dark: "#457A9F"
      active_highlight_color:
        border_foreground:
          light: "#5960D7"
          dark: "#5960D7"
    list_default_item_style:
      normal_title:
        foreground:
          light: "#2380C2"
          dark: "#2380C2"
      normal_description:
        foreground:
          light: "#457A9F"
          dark: "#457A9F"
      selected_title:
        foreground:
          light: "#585FD8"
          dark: "#585FD8"
        border_foreground:
          light: "#5960D7"
          dark: "#5960D7"
      selected_description:
        foreground:
          light: "#6C71C4"
          dark: "#6C71C4"
      dimmed_title:
        foreground:
          light: "#AB4BB0"
          dark: "#AB4BB0"
      dimmed_description:
        foreground:
          light: "#8B3D8F"
          dark: "#8B3D8F"
hostkey: >
EOF
RUN ls -alh /app
RUN ssh-keygen -N '' -f /tmp/id_hexer && cat /tmp/id_hexer | sed -e 's|\(.*\)|  \1|g' >> /app/.hexer.yaml

CMD [ "/usr/local/bin/hexer", "serve" ]