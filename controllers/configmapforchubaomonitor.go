package controllers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	cachev1alpha1 "ChubaoMonitorOperator/api/v1alpha1"
)

func ConfigmapForChubaomonitor(m *cachev1alpha1.ChubaoMonitor) *corev1.ConfigMap {
	cfg := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Configmap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "monitor-config",
			Namespace: m.Namespace,
		},
		Immutable: pointer.BoolPtr(false),
		Data: map[string]string{
			"prometheus.yml": `    # my global config
    global:
      scrape_interval:     15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
      evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
      # scrape_timeout is set to the global default (10s).


    scrape_configs:
      - job_name: 'consul'
        relabel_configs:
            - source_labels:  ["__meta_consul_service"]
              action: replace
              regex: "(.*)"
              replacement: '${1}'
              target_label: "service"
            - source_labels: ["__meta_consul_tags"]
              action: replace
              regex: ',(?:[^,]+,){0}([^=]+)=([^,]+),.*'
              replacement: '${2}'
              target_label: '${1}'
            - source_labels: ["__meta_consul_tags"]
              action: replace
              regex: ',(?:[^,]+,){1}([^=]+)=([^,]+),.*'
              replacement: '${2}'
              target_label: '${1}'
            - source_labels: ["__meta_consul_tags"]
              action: replace
              regex: ',(?:[^,]+,){2}([^=]+)=([^,]+),.*'
              replacement: '${2}'
              target_label: '${1}'
        metrics_path: /metrics
        scheme: http
        consul_sd_configs:
            - server: CONSUL_ADDRESS
              services:
                  - cfs`,

			"init.sh": `    #!/bin/sh

    host=127.0.0.1
    port=3000

    user=${GRAFANA_USERNAME:-admin}
    pass=${GRAFANA_PASSWORD:-123456}
    url="http://${user}:${pass}@${host}:${port}"
    url2="http://${host}:${port}"
    ds_url="${url2}/api/datasources"
    dashboard_url="${url2}/api/dashboard/db"

    org="chubaofs.io"

    parse_json(){
        echo "$1" | sed "s/.*\"$2\":\([^,}]*\).*/\1/"
        #echo "${1//\"/}" | sed "s/.*$2:\([^,}]*\).*/\1/"
    }

    # add org
    req=$(curl -s -X POST -H "Content-Type: application/json" -d '{"name":"${org}"}' ${url}/api/orgs)
    echo ${req}

    message=$(parse_json "${req}" "message" )

    echo "$message"

    orgid=$(parse_json "${req}" "orgId")
    # set org to admin
    curl -s -X POST ${url}/api/user/using/${orgid}

    # get token key
    #curl -s -X POST -H "Content-Type: application/json" -d '{"name":"apikeycurl", "role": "Admin"}' ${url}/api/auth/keys
    rep=$(curl -s -X POST -H "Content-Type: application/json" -d '{"name":"apikeycurl", "role": "Admin"}' ${url}/api/auth/keys)
    echo ${req}
    key=$(parse_json "${rep}" "key")
    echo "key ${key}"

    echo "add ds"
    # add datasource
    curl -s -X POST -H "Content-Type: application/json" -u ${user}:${pass} ${ds_url} -d @/grafana/ds.json
    #until curl -s -XPOST -H "Content-Type: application/json" --connect-timeout 1 -u ${userName}:${password} ${ds_url} -d @/grafana/ds.json >/dev/null; do
            #sleep 1
    #done

    # add dashboard

    #echo "add dashboard"
    curl  -X POST --insecure -H "Authorization: Bearer ${key}" -H "Content-Type: application/json" -d @/grafana/dashboard.json ${url2}/api/dashboards/db`,

			"grafana.ini": `##################### Grafana Configuration Defaults #####################
    # Do not modify this file in grafana installs


    # possible values : production, development
    app_mode = production

    # instance name, defaults to HOSTNAME environment variable value or hostname if HOSTNAME var is empty
    instance_name = ${HOSTNAME}

    #################################### Paths ##############################
    [paths]
    # Path to where grafana can store temp files, sessions, and the sqlite3 db (if that is used)
    data = data

    # Temporary files in data directory older than given duration will be removed
    temp_data_lifetime = 24h

    # Directory where grafana can store logs
    logs = data/log

    # Directory where grafana will automatically scan and look for plugins
    plugins = data/plugins

    # folder that contains provisioning config files that grafana will apply on startup and while running.
    provisioning = conf/provisioning

    #################################### Server ##############################
    [server]
    # Protocol (http, https, socket)
    protocol = http

    # The ip address to bind to, empty will bind to all interfaces
    http_addr =

    # The http port  to use
    http_port = 3000

    # The public facing domain name used to access grafana from a browser
    domain = localhost

    # Redirect to correct domain if host header does not match domain
    # Prevents DNS rebinding attacks
    enforce_domain = false

    # The full public facing url
    root_url = %(protocol)s://%(domain)s:%(http_port)s/

    # Log web requests
    router_logging = false

    # the path relative working path
    static_root_path = public

    # enable gzip
    enable_gzip = false

    # https certs & key file
    cert_file =
    cert_key =

    # Unix socket path
    socket = /tmp/grafana.sock

    #################################### Database ############################
    [database]
    # You can configure the database connection by specifying type, host, name, user and password
    # as separate properties or as on string using the url property.

    # Either "mysql", "postgres" or "sqlite3", it's your choice
    type = sqlite3
    host = 127.0.0.1:3306
    name = grafana
    user = root
    # If the password contains # or ; you have to wrap it with triple quotes. Ex """#password;"""
    password =
    # Use either URL or the previous fields to configure the database
    # Example: mysql://user:secret@host:port/database
    url =

    # Max idle conn setting default is 2
    max_idle_conn = 2

    # Max conn setting default is 0 (mean not set)
    max_open_conn =

    # Connection Max Lifetime default is 14400 (means 14400 seconds or 4 hours)
    conn_max_lifetime = 14400

    # Set to true to log the sql calls and execution times.
    log_queries =

    # For "postgres", use either "disable", "require" or "verify-full"
    # For "mysql", use either "true", "false", or "skip-verify".
    ssl_mode = disable

    ca_cert_path =
    client_key_path =
    client_cert_path =
    server_cert_name =

    # For "sqlite3" only, path relative to data_path setting
    path = grafana.db

    #################################### Session #############################
    [session]
    # Either "memory", "file", "redis", "mysql", "postgres", "memcache", default is "file"
    provider = file

    # Provider config options
    # memory: not have any config yet
    # file: session dir path, is relative to grafana data_path
    # redis: config like redis server e.g. addr=127.0.0.1:6379,pool_size=100,db=grafana
    # postgres: user=a password=b host=localhost port=5432 dbname=c sslmode=disable
    # mysql: go-sql-driver/mysql dsn config string, examples:
    #         user:password@tcp(127.0.0.1:3306)/database_name
    #         user:password@unix(/var/run/mysqld/mysqld.sock)/database_name
    # memcache: 127.0.0.1:11211


    provider_config = sessions

    # Session cookie name
    cookie_name = grafana_sess

    # If you use session in https only, default is false
    cookie_secure = false

    # Session life time, default is 86400
    session_life_time = 86400
    gc_interval_time = 86400

    # Connection Max Lifetime default is 14400 (means 14400 seconds or 4 hours)
    conn_max_lifetime = 14400

    #################################### Data proxy ###########################
    [dataproxy]

    # This enables data proxy logging, default is false
    logging = false

    #################################### Analytics ###########################
    [analytics]
    # Server reporting, sends usage counters to stats.grafana.org every 24 hours.
    # No ip addresses are being tracked, only simple counters to track
    # running instances, dashboard and error counts. It is very helpful to us.
    # Change this option to false to disable reporting.
    reporting_enabled = true

    # Set to false to disable all checks to https://grafana.com
    # for new versions (grafana itself and plugins), check is used
    # in some UI views to notify that grafana or plugin update exists
    # This option does not cause any auto updates, nor send any information
    # only a GET request to https://grafana.com to get latest versions
    check_for_updates = true

    # Google Analytics universal tracking code, only enabled if you specify an id here
    google_analytics_ua_id =

    # Google Tag Manager ID, only enabled if you specify an id here
    google_tag_manager_id =

    #################################### Security ############################
    [security]
    # default admin user, created on startup
    admin_user = admin

    # default admin password, can be changed before first start of grafana,  or in profile settings
    admin_password = admin

    # used for signing
    secret_key = SW2YcwTIb9zpOOhoPsMm

    # Auto-login remember days
    login_remember_days = 7
    cookie_username = grafana_user
    cookie_remember_name = grafana_remember

    # disable gravatar profile images
    disable_gravatar = false

    # data source proxy whitelist (ip_or_domain:port separated by spaces)
    data_source_proxy_whitelist =

    # disable protection against brute force login attempts
    disable_brute_force_login_protection = false

    #################################### Snapshots ###########################
    [snapshots]
    # snapshot sharing options
    external_enabled = true
    external_snapshot_url = https://snapshots-origin.raintank.io
    external_snapshot_name = Publish to snapshot.raintank.io

    # remove expired snapshot
    snapshot_remove_expired = true

    #################################### Dashboards ##################

    [dashboards]
    # Number dashboard versions to keep (per dashboard). Default: 20, Minimum: 1
    versions_to_keep = 20

    #################################### Users ###############################
    [users]
    # disable user signup / registration
    allow_sign_up = false

    # Allow non admin users to create organizations
    allow_org_create = false

    # Set to true to automatically assign new users to the default organization (id 1)
    auto_assign_org = true

    # Set this value to automatically add new users to the provided organization (if auto_assign_org above is set to true)
    auto_assign_org_id = 1

    # Default role new users will be automatically assigned (if auto_assign_org above is set to true)
    auto_assign_org_role = Viewer

    # Require email validation before sign up completes
    verify_email_enabled = false

    # Background text for the user field on the login page
    login_hint = email or username

    # Default UI theme ("dark" or "light")
    default_theme = dark

    # External user management
    external_manage_link_url =
    external_manage_link_name =
    external_manage_info =

    # Viewers can edit/inspect dashboard settings in the browser. But not save the dashboard.
    viewers_can_edit = false

    [auth]
    # Set to true to disable (hide) the login form, useful if you use OAuth
    disable_login_form = false

    # Set to true to disable the signout link in the side menu. useful if you use auth.proxy
    disable_signout_menu = false

    # URL to redirect the user to after sign out
    signout_redirect_url =

    #################################### Anonymous Auth ######################
    [auth.anonymous]
    # enable anonymous access
    enabled = true

    # specify organization name that should be used for unauthenticated users
    org_name = Main Org.

    # specify role for unauthenticated users
    org_role = Viewer

    #################################### Github Auth #########################
    [auth.github]
    enabled = false
    allow_sign_up = true
    client_id = some_id
    client_secret = some_secret
    scopes = user:email,read:org
    auth_url = https://github.com/login/oauth/authorize
    token_url = https://github.com/login/oauth/access_token
    api_url = https://api.github.com/user
    team_ids =
    allowed_organizations =

    #################################### GitLab Auth #########################
    [auth.gitlab]
    enabled = false
    allow_sign_up = true
    client_id = some_id
    client_secret = some_secret
    scopes = api
    auth_url = https://gitlab.com/oauth/authorize
    token_url = https://gitlab.com/oauth/token
    api_url = https://gitlab.com/api/v4
    allowed_groups =

    #################################### Google Auth #########################
    [auth.google]
    enabled = false
    allow_sign_up = true
    client_id = some_client_id
    client_secret = some_client_secret
    scopes = https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email
    auth_url = https://accounts.google.com/o/oauth2/auth
    token_url = https://accounts.google.com/o/oauth2/token
    api_url = https://www.googleapis.com/oauth2/v1/userinfo
    allowed_domains =
    hosted_domain =

    #################################### Grafana.com Auth ####################
    # legacy key names (so they work in env variables)
    [auth.grafananet]
    enabled = false
    allow_sign_up = true
    client_id = some_id
    client_secret = some_secret
    scopes = user:email
    allowed_organizations =

    [auth.grafana_com]
    enabled = false
    allow_sign_up = true
    client_id = some_id
    client_secret = some_secret
    scopes = user:email
    allowed_organizations =

    #################################### Generic OAuth #######################
    [auth.generic_oauth]
    name = OAuth
    enabled = false
    allow_sign_up = true
    client_id = some_id
    client_secret = some_secret
    scopes = user:email
    email_attribute_name = email:primary
    auth_url =
    token_url =
    api_url =
    team_ids =
    allowed_organizations =
    tls_skip_verify_insecure = false
    tls_client_cert =
    tls_client_key =
    tls_client_ca =

    #################################### Basic Auth ##########################
    [auth.basic]
    enabled = true

    #################################### Auth Proxy ##########################
    [auth.proxy]
    enabled = false
    header_name = X-WEBAUTH-USER
    header_property = username
    auto_sign_up = true
    ldap_sync_ttl = 60
    whitelist =

    #################################### Auth LDAP ###########################
    [auth.ldap]
    enabled = false
    config_file = /etc/grafana/ldap.toml
    allow_sign_up = true

    #################################### SMTP / Emailing #####################
    [smtp]
    enabled = false
    host = localhost:25
    user =
    # If the password contains # or ; you have to wrap it with triple quotes. Ex """#password;"""
    password =
    
    cert_file =
    key_file =
    skip_verify = false
    from_address = admin@grafana.localhost
    from_name = Grafana
    ehlo_identity =

    [emails]
    welcome_email_on_sign_up = false
    templates_pattern = emails/*.html

    #################################### Logging ##########################
    [log]
    # Either "console", "file", "syslog". Default is console and  file
    # Use space to separate multiple modes, e.g. "console file"
    mode = console file

    # Either "debug", "info", "warn", "error", "critical", default is "info"
    level = info

    # optional settings to set different levels for specific loggers. Ex filters = sqlstore:debug
    filters =

    # For "console" mode only
    [log.console]
    level =

    # log line format, valid options are text, console and json
    format = console

    # For "file" mode only
    [log.file]
    level =

    # log line format, valid options are text, console and json
    format = text

    # This enables automated log rotate(switch of following options), default is true
    log_rotate = true

    # Max line number of single file, default is 1000000
    max_lines = 1000000

    # Max size shift of single file, default is 28 means 1 << 28, 256MB
    max_size_shift = 28

    # Segment log daily, default is true
    daily_rotate = true

    # Expired days of log file(delete after max days), default is 7
    max_days = 7

    [log.syslog]
    level =

    # log line format, valid options are text, console and json
    format = text

    # Syslog network type and address. This can be udp, tcp, or unix. If left blank, the default unix endpoints will be used.
    network =
    address =

    # Syslog facility. user, daemon and local0 through local7 are valid.
    facility =

    # Syslog tag. By default, the process' argv[0] is used.
    tag =
    #################################### Usage Quotas ########################
    [quota]
    enabled = false

    #### set quotas to -1 to make unlimited. ####
    # limit number of users per Org.
    org_user = 10

    # limit number of dashboards per Org.
    org_dashboard = 100

    # limit number of data_sources per Org.
    org_data_source = 10

    # limit number of api_keys per Org.
    org_api_key = 10

    # limit number of orgs a user can create.
    user_org = 10

    # Global limit of users.
    global_user = -1

    # global limit of orgs.
    global_org = -1

    # global limit of dashboards
    global_dashboard = -1

    # global limit of api_keys
    global_api_key = -1

    # global limit on number of logged in users.
    global_session = -1

    #################################### Alerting ############################
    [alerting]
    # Disable alerting engine & UI features
    enabled = true
    # Makes it possible to turn off alert rule execution but alerting UI is visible
    execute_alerts = true

    # Default setting for new alert rules. Defaults to categorize error and timeouts as alerting. (alerting, keep_state)
    error_or_timeout = alerting

    # Default setting for how Grafana handles nodata or null values in alerting. (alerting, no_data, keep_state, ok)
    nodata_or_nullvalues = no_data

    # Alert notifications can include images, but rendering many images at the same time can overload the server
    # This limit will protect the server from render overloading and make sure notifications are sent out quickly
    concurrent_render_limit = 5

    #################################### Explore #############################
    [explore]
    # Enable the Explore section
    enabled = false

    #################################### Internal Grafana Metrics ############
    # Metrics available at HTTP API Url /metrics
    [metrics]
    enabled           = true
    interval_seconds  = 10

    # Send internal Grafana metrics to graphite
    [metrics.graphite]
    # Enable by setting the address setting (ex localhost:2003)
    address =
    prefix = prod.grafana.%(instance_name)s.

    [grafana_net]
    url = https://grafana.com

    [grafana_com]
    url = https://grafana.com

    #################################### Distributed tracing ############
    [tracing.jaeger]
    # jaeger destination (ex localhost:6831)
    address =
    # tag that will always be included in when creating new spans. ex (tag1:value1,tag2:value2)
    always_included_tag =
    # Type specifies the type of the sampler: const, probabilistic, rateLimiting, or remote
    sampler_type = const
    # jaeger samplerconfig param
    # for "const" sampler, 0 or 1 for always false/true respectively
    # for "probabilistic" sampler, a probability between 0 and 1
    # for "rateLimiting" sampler, the number of spans per second
    # for "remote" sampler, param is the same as for "probabilistic"
    # and indicates the initial sampling rate before the actual one
    # is received from the mothership
    sampler_param = 1

    #################################### External Image Storage ##############
    [external_image_storage]
    # You can choose between (s3, webdav, gcs, azure_blob, local)
    provider =

    [external_image_storage.s3]
    bucket_url =
    bucket =
    region =
    path =
    access_key =
    secret_key =

    [external_image_storage.webdav]
    url =
    username =
    password =
    public_url =

    [external_image_storage.gcs]
    key_file =
    bucket =
    path =

    [external_image_storage.azure_blob]
    account_name =
    account_key =
    container_name =

    [external_image_storage.local]
    # does not require any configuration

    [rendering]
    # Options to configure external image rendering server like https://github.com/grafana/grafana-image-renderer
    server_url =
    callback_url =`,

			"datasource.yml": `    apiVersion: 1

    # list of datasources that should be deleted from the database
    deleteDatasources:
      - name: Prometheus
        orgId: 1

    # list of datasources to insert/update depending
    # whats available in the database
    datasources:
      # <string, required> name of the datasource. Required
    - name: Prometheus
      # <string, required> datasource type. Required
      type: prometheus
      # <string, required> access mode. direct or proxy. Required
      access: proxy
      # <int> org id. will default to orgId 1 if not specified
      orgId: 1
      # <string> url
      url: PROMETHEUS_URL
      # <string> database password, if used
      password:
      # <string> database user, if used
      user:
      # <string> database name, if used
      database:
      # <bool> enable/disable basic auth
      basicAuth: false
      # <string> basic auth username
      basicAuthUser: admin
      # <string> basic auth password
      basicAuthPassword: 123456
      # <bool> enable/disable with credentials headers
      withCredentials:
      # <bool> mark as default datasource. Max one per org
      isDefault: true
      # <map> fields that will be converted to json and stored in json_data
      jsonData:
        graphiteVersion: "1.1"
        tlsAuth: false
        tlsAuthWithCACert: false
        # <string> json object of data that will be encrypted.
      secureJsonData:
        tlsCACert: "..."
        tlsClientCert: "..."
        tlsClientKey: "..."
      version: 1
      # <bool> allow users to edit datasources from the UI.
      editable: true`,

			"dashboard.yml": `    apiVersion: 1

    providers:
    - name: 'ChubaoFS'
      orgId: 1
      folder: ''
      type: file
      disableDeletion: false
      editable: true
      options:
        path: /etc/grafana/provisioning/dashboards`,

			"chubaofs.json": `     {
      "annotations": {
        "list": [
          {
            "builtIn": 1,
            "datasource": "-- Grafana --",
            "enable": true,
            "hide": true,
            "iconColor": "rgba(0, 211, 255, 1)",
            "name": "Annotations & Alerts",
            "type": "dashboard"
          }
        ]
      },
	  "editable": true,
      "gnetId": null,
      "graphTooltip": 0,
      "id": 1,
      "iteration": 1593328519495,
      "links": [],
      "panels": [
        {
          "collapsed": false,
          "datasource": null,
          "gridPos": {
            "h": 1,
            "w": 24,
            "x": 0,
            "y": 0
          },
          "id": 40,
          "panels": [],
          "title": "Cluster",
          "type": "row"
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
		  "datasource": null,
          "description": "master node total count of cluster",
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 4,
            "w": 3,
            "x": 0,
            "y": 1
          },
          "id": 38,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "lineColor": "rgb(31, 120, 193)",
            "show": true
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "count(up{cluster=\"$cluster\", app=\"$app\", role=\"master\"} > 0)",
              "format": "time_series",
              "intervalFactor": 1,
              "refId": "A"
            }
          ],
          "thresholds": "",
          "title": "MasterCount",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": null,
          "description": "Metanode total count for current selected cluster",
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 4,
            "w": 3,
            "x": 3,
            "y": 1
          },
          "id": 42,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "lineColor": "rgb(31, 120, 193)",
            "show": true
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "count(up{cluster=\"$cluster\", app=\"$app\", role=\"metanode\"} > 0)",
              "format": "time_series",
              "intervalFactor": 1,
              "refId": "A"
            }
          ],
          "thresholds": "",
          "title": "MetaNodeCount",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": null,
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 4,
            "w": 3,
            "x": 6,
            "y": 1
          },
          "id": 41,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "lineColor": "rgb(31, 120, 193)",
            "show": true
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "count(up{cluster=\"$cluster\", app=\"$app\", role=\"dataNode\"} > 0)",
              "format": "time_series",
              "intervalFactor": 1,
              "refId": "A"
            }
          ],
          "thresholds": "",
          "title": "DataNodeCount",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": null,
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 4,
            "w": 3,
            "x": 9,
            "y": 1
          },
          "id": 86,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
		  "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "lineColor": "rgb(31, 120, 193)",
            "show": true
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "count(go_info{cluster=\"$cluster\", app=\"$app\", role=\"objectnode\"})",
              "format": "time_series",
              "intervalFactor": 1,
              "refId": "A"
            }
          ],
          "thresholds": "",
          "title": "ObjectNodeCount",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": null,
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 4,
            "w": 3,
            "x": 12,
            "y": 1
          },
          "id": 287,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "lineColor": "rgb(31, 120, 193)",
            "show": true
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "count(up{cluster=\"$cluster\", app=\"$app\", role=\"fuseclient\"} > 0)",
              "format": "time_series",
              "intervalFactor": 1,
              "refId": "A"
            }
          ],
          "thresholds": "",
          "title": "ClientCount",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": null,
          "description": "Cluster volume total count",
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 4,
            "w": 3,
            "x": 15,
            "y": 1
          },
          "id": 93,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
		  "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "lineColor": "rgb(31, 120, 193)",
            "show": true
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "sum(cfs_master_vol_count{cluster=~\"$cluster\", instance=\"$master_instance\"})",
              "format": "time_series",
              "intervalFactor": 1,
              "refId": "A"
            }
          ],
          "thresholds": "",
          "title": "VolumeCount",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": null,
          "description": "Chubaofs process start time of selected instance",
          "format": "dateTimeFromNow",
          "gauge": {
            "maxValue": 100,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 4,
            "w": 3,
            "x": 18,
            "y": 1
          },
          "id": 113,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "lineColor": "rgb(31, 120, 193)",
            "show": true
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "process_start_time_seconds{cluster=\"$cluster\", app=\"$app\", instance=~\"$instance\"}*1000",
              "format": "time_series",
              "intervalFactor": 1,
              "refId": "A"
            }
          ],
          "thresholds": "",
          "title": "ProcessStartTime",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "N/A",
              "value": "null"
            }
          ],
          "valueName": "current"
        },
        {
          "datasource": null,
          "gridPos": {
            "h": 6,
            "w": 6,
            "x": 0,
            "y": 5
          },
          "id": 177,
          "options": {
            "displayMode": "gradient",
            "fieldOptions": {
              "calcs": [
                "lastNotNull"
              ],
              "defaults": {
                "mappings": [],
                "max": 300000,
                "min": 0,
                "thresholds": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "#EAB839",
                    "value": 200000
                  },
                  {
                    "color": "red",
                    "value": 300000
                  }
                ],
                "unit": "gbytes"
              },
              "override": {},
              "values": false
            },
            "orientation": "horizontal"
          },
          "pluginVersion": "6.4.4",
          "targets": [
            {
              "expr": "cfs_master_dataNodes_total_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"}",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
			  "legendFormat": "TotalSize",
              "refId": "B"
            },
            {
              "expr": "cfs_master_dataNodes_used_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"}",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "UsedSize",
              "refId": "A"
            },
            {
              "expr": "cfs_master_dataNodes_increased_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"}",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "IncreasedSize",
              "refId": "C"
            }
          ],
          "title": "DatanodeSize",
          "type": "bargauge"
        },
        {
          "datasource": null,
          "description": "metanode size of total, used and increased.",
          "gridPos": {
            "h": 6,
            "w": 6,
            "x": 6,
            "y": 5
          },
          "id": 92,
          "options": {
            "displayMode": "gradient",
            "fieldOptions": {
              "calcs": [
                "last"
              ],
              "defaults": {
                "mappings": [],
                "max": 5,
                "min": 0,
                "thresholds": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "#EAB839",
                    "value": 3
                  },
                  {
                    "color": "red",
                    "value": 5
                  }
                ],
                "unit": "gbytes"
              },
              "override": {},
              "values": false
            },
            "orientation": "horizontal"
          },
          "pluginVersion": "6.4.4",
          "targets": [
            {
              "expr": "sum(cfs_master_metaNodes_total_GB{cluster=~\"$cluster\", instance=\"$master_instance\"})",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "TotalSize",
              "refId": "B"
            },
            {
              "expr": "sum(cfs_master_metaNodes_used_GB{cluster=~\"$cluster\", instance=\"$master_instance\"} )",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "UsedSize",
              "refId": "A"
            },
            {
              "expr": "sum(cfs_master_metaNodes_increased_GB{cluster=~\"$cluster\", instance=\"$master_instance\"})",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "IncreasedSize",
              "refId": "C"
            }
          ],
          "title": "MetanodeSize",
          "type": "bargauge"
        },
        {
          "datasource": null,
          "description": "Top K total size of all volume",
          "gridPos": {
            "h": 6,
            "w": 9,
            "x": 12,
            "y": 5
          },
          "id": 175,
          "options": {
            "displayMode": "gradient",
            "fieldOptions": {
              "calcs": [
                "last"
              ],
              "defaults": {
                "mappings": [
                  {
                    "op": "=",
                    "text": "N/A",
                    "type": 2,
                    "value": "null"
                  }
                ],
                "max": 100,
                "min": 0,
                "nullValueMode": "connected",
                "thresholds": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "#EAB839",
                    "value": 50
                  },
                  {
                    "color": "red",
                    "value": 100
                  }
                ],
                "unit": "bytes"
              },
              "override": {},
              "values": true
            },
            "orientation": "horizontal"
          },
          "pluginVersion": "6.4.4",
          "targets": [
            {
              "expr": "topk($topk, cfs_master_vol_total_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"})",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "{{volName}}",
              "refId": "A"
            }
          ],
          "title": "VolumeTotalSize",
          "type": "bargauge"
        },
        {
          "datasource": null,
          "description": "Datanode used size ratio with total size",
          "gridPos": {
            "h": 6,
            "w": 3,
            "x": 0,
            "y": 11
          },
          "id": 281,
          "options": {
            "fieldOptions": {
              "calcs": [
                "last"
              ],
              "defaults": {
                "mappings": [],
                "max": 1,
                "min": 0,
                "thresholds": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "#EAB839",
                    "value": 0.5
                  },
                  {
                    "color": "red",
                    "value": 0.8
                  }
                ],
                "unit": "percentunit"
              },
              "override": {},
              "values": false
            },
            "orientation": "horizontal",
            "showThresholdLabels": false,
            "showThresholdMarkers": true
          },
          "pluginVersion": "6.4.4",
          "targets": [
            {
              "expr": "cfs_master_dataNodes_used_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"}/ cfs_master_dataNodes_total_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"}",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "DataNodeUsedRatio",
              "refId": "D"
            }
          ],
          "title": "DataNodeUsedRatio",
          "type": "gauge"
        },
        {
          "datasource": null,
          "description": "Metanode used size ratio with total size",
          "gridPos": {
            "h": 6,
            "w": 3,
            "x": 3,
            "y": 11
          },
          "id": 282,
          "options": {
            "fieldOptions": {
              "calcs": [
                "last"
              ],
              "defaults": {
                "mappings": [],
                "max": 1,
                "min": 0,
                "thresholds": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "#EAB839",
                    "value": 0.5
                  },
                  {
                    "color": "red",
                    "value": 0.8
                  }
                ],
                "unit": "percentunit"
              },
              "override": {},
              "values": false
            },
            "orientation": "horizontal",
            "showThresholdLabels": false,
            "showThresholdMarkers": true
          },
          "pluginVersion": "6.4.4",
          "targets": [
            {
              "expr": "cfs_master_metaNodes_used_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"} / cfs_master_metaNodes_total_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"}",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "MetaNodeUsedRatio",
			  "refId": "A"
            }
          ],
          "title": "MetaNodeUsedRatio",
          "type": "gauge"
        },
        {
          "datasource": null,
          "description": "Datanode inactive count",
          "gridPos": {
            "h": 6,
            "w": 3,
            "x": 6,
            "y": 11
          },
          "id": 181,
          "options": {
            "fieldOptions": {
              "calcs": [
                "lastNotNull"
              ],
              "defaults": {
                "mappings": [],
                "max": 100,
                "min": 0,
                "thresholds": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "red",
                    "value": 80
                  }
                ]
              },
              "override": {},
              "values": false
            },
            "orientation": "auto",
            "showThresholdLabels": false,
            "showThresholdMarkers": true
          },
          "pluginVersion": "6.4.4",
          "targets": [
            {
              "expr": "cfs_master_dataNodes_inactive{app=\"$app\", cluster=\"$cluster\", instance=\"$master_instance\"}",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "sum",
              "refId": "A"
            }
          ],
          "title": "DataNodeInactive",
          "type": "gauge"
        },
        {
          "datasource": null,
          "description": "Metanode inactive count",
          "gridPos": {
            "h": 6,
            "w": 3,
            "x": 9,
            "y": 11
          },
          "id": 180,
          "options": {
            "fieldOptions": {
              "calcs": [
                "lastNotNull"
              ],
              "defaults": {
                "mappings": [],
                "max": 100,
                "min": 0,
                "thresholds": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "red",
                    "value": 80
                  }
                ]
              },
              "override": {},
              "values": false
            },
            "orientation": "auto",
            "showThresholdLabels": false,
            "showThresholdMarkers": true
          },
          "pluginVersion": "6.4.4",
          "targets": [
            {
              "expr": "cfs_master_metaNodes_inactive{app=\"$app\", cluster=\"$cluster\", instance=\"$master_instance\"}",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "sum",
              "refId": "A"
            }
          ],
          "title": "MetaNodesInactive",
          "type": "gauge"
        },
        {
          "cacheTimeout": null,
          "colorBackground": false,
          "colorValue": false,
          "colors": [
            "#299c46",
            "rgba(237, 129, 40, 0.89)",
            "#d44a3a"
          ],
          "datasource": null,
          "description": "Disk Error count",
          "format": "none",
          "gauge": {
            "maxValue": 100,
            "minValue": 0,
            "show": false,
            "thresholdLabels": false,
            "thresholdMarkers": true
          },
          "gridPos": {
            "h": 6,
            "w": 3,
            "x": 12,
            "y": 11
          },
          "id": 179,
          "interval": null,
          "links": [],
          "mappingType": 1,
          "mappingTypes": [
            {
              "name": "value to text",
              "value": 1
            },
            {
              "name": "range to text",
              "value": 2
            }
          ],
          "maxDataPoints": 100,
          "nullPointMode": "connected",
          "nullText": null,
          "options": {},
          "pluginVersion": "6.4.4",
          "postfix": "",
          "postfixFontSize": "50%",
          "prefix": "",
          "prefixFontSize": "50%",
          "rangeMaps": [
            {
              "from": "null",
              "text": "N/A",
              "to": "null"
            }
          ],
          "sparkline": {
            "fillColor": "rgba(31, 118, 189, 0.18)",
            "full": false,
            "lineColor": "rgb(31, 120, 193)",
            "show": true,
            "ymax": null,
            "ymin": null
          },
          "tableColumn": "",
          "targets": [
            {
              "expr": "delta(cfs_master_disk_error{app=\"$app\", cluster=\"$cluster\", instance=\"$master_instance\"}[1m])",
              "format": "time_series",
              "intervalFactor": 1,
              "refId": "A"
            }
          ],
          "thresholds": "",
          "title": "DiskError",
          "type": "singlestat",
          "valueFontSize": "80%",
          "valueMaps": [
            {
              "op": "=",
              "text": "0",
              "value": "null"
            }
          ],
          "valueName": "current"
        },
        {
          "datasource": null,
          "description": "top k volume used ratio ",
          "gridPos": {
            "h": 6,
            "w": 6,
            "x": 15,
            "y": 11
          },
          "id": 280,
          "options": {
            "displayMode": "gradient",
            "fieldOptions": {
              "calcs": [
                "last"
              ],
              "defaults": {
                "mappings": [
                  {
                    "op": "=",
                    "text": "N/A",
                    "type": 1,
                    "value": "null"
                  }
                ],
                "max": 1,
                "min": 0,
                "nullValueMode": "connected",
                "thresholds": [
                  {
                    "color": "green",
                    "value": null
                  },
                  {
                    "color": "#EAB839",
                    "value": 0.5
                  },
                  {
                    "color": "red",
                    "value": 0.8
                  }
                ],
                "unit": "percentunit"
              },
              "override": {},
              "values": false
            },
            "orientation": "horizontal"
          },
          "pluginVersion": "6.4.4",
          "targets": [
            {
              "expr": "topk($topk, cfs_master_vol_used_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"}/cfs_master_vol_total_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"})",
              "format": "time_series",
              "instant": true,
              "intervalFactor": 1,
              "legendFormat": "{{volName}}",
              "refId": "A"
            }
          ],
          "title": "VolumeUsedRatio",
          "type": "bargauge"
        },
        {
          "collapsed": false,
          "datasource": null,
          "gridPos": {
            "h": 1,
            "w": 24,
            "x": 0,
            "y": 17
          },
          "id": 184,
          "panels": [],
          "title": "Volume",
          "type": "row"
        },
        {
          "aliasColors": {},
          "bars": false,
          "dashLength": 10,
          "dashes": false,
          "datasource": null,
          "description": "top k volume used size",
          "fill": 1,
          "fillGradient": 0,
          "gridPos": {
            "h": 6,
            "w": 6,
            "x": 0,
            "y": 18
          },
          "id": 187,
          "legend": {
            "avg": false,
            "current": false,
            "max": false,
            "min": false,
            "show": true,
            "total": false,
            "values": false
          },
          "lines": true,
          "linewidth": 1,
          "nullPointMode": "null",
          "options": {
            "dataLinks": []
          },
          "percentage": false,
          "pluginVersion": "6.4.4",
          "pointradius": 5,
          "points": false,
          "renderer": "flot",
          "seriesOverrides": [
            {
              "alias": "/.*UsedRatio.*/",
              "yaxis": 2
            }
          ],
          "spaceLength": 10,
          "stack": false,
          "steppedLine": false,
          "targets": [
            {
              "expr": "topk($topk, cfs_master_vol_used_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"})",
              "format": "time_series",
              "intervalFactor": 1,
              "legendFormat": "{{volName}}",
			  "refId": "A"
            }
          ],
          "thresholds": [],
          "timeFrom": null,
          "timeRegions": [],
          "timeShift": null,
          "title": "VolumeUsedSize",
          "tooltip": {
            "shared": true,
            "sort": 0,
            "value_type": "individual"
          },
          "type": "graph",
          "xaxis": {
            "buckets": null,
            "mode": "time",
            "name": null,
            "show": true,
            "values": []
          },
          "yaxes": [
            {
              "format": "bytes",
              "logBase": 10,
              "show": true
            },
            {
              "format": "percentunit",
              "logBase": 1,
              "show": true
            }
          ],
          "yaxis": {
            "align": false,
            "alignLevel": null
          }
        },
        {
          "aliasColors": {},
          "bars": false,
          "dashLength": 10,
          "dashes": false,
          "datasource": null,
          "description": "top k volume used ratio",
          "fill": 1,
          "fillGradient": 0,
          "gridPos": {
            "h": 6,
            "w": 6,
            "x": 6,
            "y": 18
          },
          "id": 188,
          "legend": {
            "avg": false,
            "current": false,
            "max": false,
            "min": false,
            "show": true,
            "total": false,
            "values": false
          },
          "lines": true,
          "linewidth": 1,
          "nullPointMode": "null",
          "options": {
            "dataLinks": []
          },
          "percentage": false,
          "pluginVersion": "6.4.4",
          "pointradius": 5,
          "points": false,
          "renderer": "flot",
          "seriesOverrides": [
            {
              "alias": "/.*UsedRatio.*/",
              "yaxis": 2
            }
          ],
          "spaceLength": 10,
          "stack": false,
          "steppedLine": false,
          "targets": [
            {
              "expr": "topk($topk, cfs_master_vol_used_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"}/cfs_master_vol_total_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"})",
              "format": "time_series",
              "intervalFactor": 1,
              "legendFormat": "{{volName}}",
              "refId": "B"
            }
          ],
          "thresholds": [],
          "timeFrom": null,
          "timeRegions": [],
          "timeShift": null,
          "title": "VolumeUsedRatio",
          "tooltip": {
            "shared": true,
            "sort": 2,
            "value_type": "individual"
          },
          "type": "graph",
          "xaxis": {
            "buckets": null,
            "mode": "time",
            "name": null,
            "show": true,
            "values": []
          },
          "yaxes": [
            {
              "format": "percentunit",
              "logBase": 1,
              "show": true
            },
            {
              "format": "percentunit",
              "logBase": 1,
              "show": true
            }
          ],
          "yaxis": {
            "align": false,
            "alignLevel": null
          }
        },
        {
          "aliasColors": {},
          "bars": false,
          "dashLength": 10,
          "dashes": false,
          "datasource": null,
          "description": "top k volume size rate",
          "fill": 1,
          "fillGradient": 0,
          "gridPos": {
            "h": 6,
            "w": 6,
            "x": 12,
            "y": 18
          },
          "id": 182,
          "legend": {
            "avg": false,
            "current": false,
            "max": false,
            "min": false,
            "show": true,
            "total": false,
            "values": false
          },
          "lines": true,
          "linewidth": 1,
          "nullPointMode": "null",
          "options": {
            "dataLinks": []
          },
          "percentage": false,
          "pointradius": 5,
          "points": false,
          "renderer": "flot",
          "seriesOverrides": [
            {
              "alias": "/.*UsedRatio.*/",
              "yaxis": 2
            }
          ],
          "spaceLength": 10,
          "stack": false,
          "steppedLine": false,
          "targets": [
            {
              "expr": "topk($topk, rate(cfs_master_vol_used_GB{app=\"$app\",cluster=\"$cluster\", instance=\"$master_instance\"}[1m]))",
              "format": "time_series",
              "intervalFactor": 1,
              "legendFormat": "{{volName}}",
              "refId": "C"
            }
          ],
          "thresholds": [],
          "timeFrom": null,
          "timeRegions": [],
          "timeShift": null,
          "title": "VolumeSizeRate",
          "tooltip": {
            "shared": true,
            "sort": 2,
            "value_type": "individual"
          },
          "type": "graph",
          "xaxis": {
            "buckets": null,
            "mode": "time",
            "name": null,
            "show": true,
            "values": []
          },
          "yaxes": [
            {
              "format": "Bps",
              "logBase": 1,
              "show": true
            },
            {
              "format": "percentunit",
              "logBase": 1,
              "show": true
            }
          ],
          "yaxis": {
            "align": false,
            "alignLevel": null
          }
        },
        {
          "aliasColors": {},
          "bars": false,
          "dashLength": 10,
          "dashes": false,
          "datasource": null,
          "description": "Selected volume used/total size and used ratio",
          "fill": 1,
          "fillGradient": 0,
          "gridPos": {
            "h": 6,
            "w": 6,
            "x": 18,
            "y": 18
          },
          "id": 186,
          "legend": {
            "avg": false,
            "current": false,
            "max": false,
            "min": false,
            "show": true,
            "total": false,
            "values": false
          },
          "lines": true,
          "linewidth": 1,
          "nullPointMode": "null",
          "options": {
            "dataLinks": []
          },
          "percentage": false,
          "pointradius": 5,
          "points": false,
          "renderer": "flot",
          "seriesOverrides": [
            {
              "alias": "/.*UsedRatio.*/",
              "yaxis": 2
            }
          ],
          "spaceLength": 10,
          "stack": false,
          "steppedLine": false,
          "targets": [
            {
              "expr": "cfs_master_vol_used_GB{app=\"$app\",cluster=\"$cluster\", volName=\"$vol\", instance=\"$master_instance\"}",
              "format": "time_series",
              "intervalFactor": 1,
              "legendFormat": "UsedSize",
              "refId": "C"
            },
            {
              "expr": "cfs_master_vol_used_GB{app=\"$app\",cluster=\"$cluster\", volName=\"$vol\", instance=\"$master_instance\"}/cfs_master_vol_total_GB{app=\"$app\",cluster=\"$cluster\", volName=\"$vol\", instance=\"$master_instance\"}",
              "format": "time_series",
              "intervalFactor": 1,
              "legendFormat": "UsedRatio",
              "refId": "A"
            }
          ],
          "thresholds": [],
          "timeFrom": null,
          "timeRegions": [],
          "timeShift": null,
          "title": "[[vol]]",
          "tooltip": {
            "shared": true,
            "sort": 0,
            "value_type": "individual"
          },
          "type": "graph",
          "xaxis": {
            "buckets": null,
            "mode": "time",
            "name": null,
            "show": true,
			"values": []
          },
          "yaxes": [
            {
              "format": "bytes",
              "logBase": 1,
              "show": true
            },
            {
              "format": "percentunit",
              "logBase": 1,
              "show": true
            }
          ],
          "yaxis": {
            "align": false,
            "alignLevel": null
          }
        },
        {
          "collapsed": true,
          "datasource": null,
          "gridPos": {
            "h": 1,
            "w": 24,
            "x": 0,
            "y": 24
          },
          "id": 34,
          "panels": [
            {
              "aliasColors": {},
              "bars": true,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "invalide master nodes",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 7,
                "x": 0,
                "y": 25
              },
              "id": 162,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": false,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "master_count",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, 1 - up{cluster=\"$cluster\", role=\"master\"} > 0 )",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "[[cluster]]_master_{{instance}}",
                  "refId": "C"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "master_nodes_invalide",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": true,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "invalide meta nodes",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 7,
                "x": 7,
                "y": 25
              },
              "id": 163,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": true,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, 1 - up{cluster=\"$cluster\", role=\"metanode\"} > 0 )",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{role}}_{{instance}}",
                  "refId": "C"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[cluster]]_metanode_inactive",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": true,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "invalide data nodes",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 7,
                "x": 0,
                "y": 32
              },
              "id": 164,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": true,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, 1 - up{cluster=\"$cluster\", role=\"dataNode\"}  > 0)",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{role}}_{{instance}}",
                  "refId": "C"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "datanode_inactive",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": true,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "invalide fuse clients",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 7,
                "x": 7,
                "y": 32
              },
              "id": 165,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": true,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, 1 - up{app=\"$app\", cluster=\"$cluster\", role=\"fuseclient\"}  > 0)",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{role}}_{{instance}}",
                  "refId": "C"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "fuseclient_inactive",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            }
          ],
          "title": "Master",
          "type": "row"
        },
        {
          "collapsed": true,
          "datasource": null,
          "gridPos": {
            "h": 1,
            "w": 24,
            "x": 0,
            "y": 25
          },
          "id": 36,
          "panels": [
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "metanode op time",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 8,
                "w": 24,
                "x": 0,
                "y": 40
              },
              "id": 117,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": false,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "mnOp",
              "repeatDirection": "h",
              "scopedVars": {
                "mnOp": {
                  "selected": true,
                  "text": "OpMetaInodeGet",
                  "value": "OpMetaInodeGet"
                }
              },
              "seriesOverrides": [
                {
                  "alias": "/sum/",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, cfs_metanode_[[mnOp]]{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "sum( cfs_metanode_[[mnOp]]{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
			   "title": "[[mnOp]]:Time",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "ns",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ns",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "metanode op count per second",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 8,
                "w": 24,
                "x": 0,
                "y": 48
              },
              "id": 205,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": false,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "mnOp",
              "repeatDirection": "h",
              "scopedVars": {
                "mnOp": {
                  "selected": true,
                  "text": "OpMetaInodeGet",
                  "value": "OpMetaInodeGet"
                }
              },
              "seriesOverrides": [
                {
                  "alias": "/sum/",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, cfs_metanode_[[mnOp]]_count{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "sum( cfs_metanode_[[mnOp]]_count{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[mnOp]]:Ops",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            }
          ],
          "title": "Metanode",
          "type": "row"
        },
        {
          "collapsed": true,
          "datasource": null,
          "gridPos": {
            "h": 1,
            "w": 24,
            "x": 0,
            "y": 26
          },
          "id": 27,
          "panels": [
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "Datanode op time",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 24,
                "x": 0,
                "y": 57
              },
              "id": 146,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": false,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "dnOp",
              "repeatDirection": "h",
              "scopedVars": {
                "dnOp": {
                  "selected": true,
                  "text": "OpCreateExtent",
                  "value": "OpCreateExtent"
                }
              },
              "seriesOverrides": [
                {
                  "alias": "sum",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, cfs_dataNode_[[dnOp]]{app=~\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "P"
                },
                {
                  "expr": "sum(cfs_dataNode_[[dnOp]]{app=~\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[dnOp]]:Time",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "ns",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ns",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "Datanode op count per second",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 24,
                "x": 0,
                "y": 64
              },
              "id": 200,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "dnOp",
              "repeatDirection": "h",
              "scopedVars": {
                "dnOp": {
                  "selected": true,
                  "text": "OpCreateExtent",
                  "value": "OpCreateExtent"
                }
              },
              "seriesOverrides": [
                {
                  "alias": "/sum/",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, rate(cfs_dataNode_[[dnOp]]_count{app=~\"$app\", cluster=\"$cluster\"}[1m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "P"
                },
                {
                  "expr": "sum(rate(cfs_dataNode_[[dnOp]]_count{app=~\"$app\", cluster=\"$cluster\"}[1m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[dnOp]]:Ops",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            }
          ],
          "title": "Datanode",
          "type": "row"
        },
        {
          "collapsed": true,
          "datasource": null,
          "gridPos": {
            "h": 1,
            "w": 24,
            "x": 0,
            "y": 27
          },
          "id": 284,
          "panels": [
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "Object node op time",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 24,
                "x": 0,
                "y": 42
              },
              "id": 285,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "obOp",
              "repeatDirection": "h",
              "scopedVars": {
                "obOp": {
                  "selected": true,
                  "text": "ListObjects",
                  "value": "ListObjects"
                }
              },
              "seriesOverrides": [
                {
                  "alias": "sum",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, cfs_objectnode_action_[[obOp]]{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "P"
                },
                {
                  "expr": "sum(cfs_objectnode_action_[[obOp]]{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[obOp]]:Time",
			  "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "ns",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ns",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "Object node op count per second",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 24,
                "x": 0,
                "y": 49
              },
              "id": 286,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "obOp",
              "repeatDirection": "h",
              "scopedVars": {
                "obOp": {
                  "selected": true,
                  "text": "ListObjects",
                  "value": "ListObjects"
                }
              },
              "seriesOverrides": [
                {
                  "alias": "/sum/",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, rate(cfs_objectnode_action_[[obOp]]_count{app=~\"$app\", cluster=\"$cluster\"}[1m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "P"
                },
                {
                  "expr": "sum(rate(cfs_objectnode_action_[[obOp]]_count{app=~\"$app\", cluster=\"$cluster\"}[1m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[obOp]]:Ops",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            }
          ],
          "title": "ObjectNode",
          "type": "row"
        },
        {
          "collapsed": true,
          "datasource": null,
          "gridPos": {
            "h": 1,
            "w": 24,
            "x": 0,
            "y": 28
          },
          "id": 66,
          "panels": [
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "fuse client use meta sdk op time",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 24,
                "x": 0,
                "y": 43
              },
              "id": 64,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "mnOp",
              "repeatDirection": "h",
              "scopedVars": {
                "mnOp": {
                  "selected": true,
                  "text": "OpMetaInodeGet",
                  "value": "OpMetaInodeGet"
                }
              },
              "seriesOverrides": [
                {
                  "alias": "/sum/",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, cfs_rolefuseclient_[[mnOp]]{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "K"
                },
                {
                  "expr": "sum(cfs_fuseclient_[[mnOp]]{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[mnOp]]:Time",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "ns",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "fuse client use meta sdk op count per second",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 24,
                "x": 0,
                "y": 50
              },
              "id": 212,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "mnOp",
              "repeatDirection": "h",
              "scopedVars": {
                "mnOp": {
                  "selected": true,
                  "text": "OpMetaInodeGet",
                  "value": "OpMetaInodeGet"
                }
              },
              "seriesOverrides": [
                {
                  "alias": "/sum/",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, rate(cfs_fuseclient_[[mnOp]]{app=\"$app\", cluster=\"$cluster\"}[1m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "K"
                },
                {
                  "expr": "sum(rate(cfs_fuseclient_[[mnOp]]{app=\"$app\", cluster=\"$cluster\"}[1m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[mnOp]]:Ops",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "fuse client op time",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 24,
                "x": 0,
                "y": 57
              },
              "id": 215,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "fuseOp",
              "repeatDirection": "h",
              "scopedVars": {
                "fuse2Op": {
                  "selected": true,
                  "text": "mkdir",
                  "value": "mkdir"
                },
                "fuseOp": {
                  "selected": true,
                  "text": "mkdir",
                  "value": "mkdir"
                }
              },
              "seriesOverrides": [
                {
                  "alias": "/sum/",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, cfs_fuseclient_[[fuseOp]]{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "K"
                },
                {
                  "expr": "sum(cfs_fuseclient_[[fuseOp]]{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[fuseOp]]:Time",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "ns",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": null,
              "description": "fuse op count per second",
              "fill": 1,
              "fillGradient": 0,
              "gridPos": {
                "h": 7,
                "w": 24,
                "x": 0,
                "y": 64
              },
              "id": 216,
              "legend": {
                "avg": false,
                "current": false,
                "max": false,
                "min": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 1,
              "maxPerRow": 4,
              "nullPointMode": "null",
              "options": {
                "dataLinks": []
              },
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "repeat": "fuseOp",
              "repeatDirection": "h",
              "scopedVars": {
                "fuseOp": {
                  "selected": true,
                  "text": "mkdir",
                  "value": "mkdir"
                }
              },
              "seriesOverrides": [
                {
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "topk($topk, cfs_fuseclient_[[fuseOp]]_count{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{instance}}",
                  "refId": "K"
                },
                {
                  "expr": "sum(cfs_fuseclient_[[fuseOp]]_count{app=\"$app\", cluster=\"$cluster\"})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sum",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeRegions": [],
              "timeShift": null,
              "title": "[[fuseOp]]:Ops",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
				"values": []
              },
              "yaxes": [
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ops",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            }
          ],
          "title": "FuseClient",
          "type": "row"
        },
        {
          "collapsed": true,
          "datasource": null,
          "gridPos": {
            "h": 1,
            "w": 24,
            "x": 0,
            "y": 29
          },
          "id": 60,
          "panels": [
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "y": 25
              },
              "id": 61,
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "topk($topk, rate(go_goroutines{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}[1m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{role}}_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "topk($topk,go_threads{app=\"$app\", cluster=\"$cluster\", role=\"$role\"})",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "go_threads_{{instance}}",
                  "refId": "B"
                }
              ],
              "title": "go_goroutines",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 7,
                "y": 25
              },
              "id": 169,
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "topk($topk, rate(go_threads{app=\"$app\", cluster=\"$cluster\"}[1m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "{{role}}_{{instance}}",
                  "refId": "B"
                }
              ],
              "title": "go_threads",
              "tooltip": {
                "shared": true,
                "sort": 2,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 14,
                "y": 25
              },
              "id": 63,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "gc_rate",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "go_gc_duration_seconds{instance=~\"$instance\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "seconds_{{quantile}}",
                  "refId": "A"
                },
                {
                  "expr": "rate(go_gc_duration_seconds_count{instance=~\"$instance\"}[1m])",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "gc_rate",
                  "refId": "B"
                }
              ],
              "title": "gc_duration",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "s",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "y": 31
              },
              "id": 109,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "gc_rate",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "process_resident_memory_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\" }",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "resident_memory_bytes_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "process_virtual_memory_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "virtual_memory_bytes_{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "process_virtual_memory_max_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "virtual_memory_max_bytes_{{instance}}",
                  "refId": "C"
                }
              ],
              "title": "process_resident_memory",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "decbytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 7,
                "y": 31
              },
              "id": 166,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "gc_rate",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "process_resident_memory_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\" }",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "resident_memory_bytes_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "process_virtual_memory_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "virtual_memory_bytes_{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "process_virtual_memory_max_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "virtual_memory_max_bytes_{{instance}}",
                  "refId": "C"
                }
              ],
              "title": "process_virtual_memory",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "decbytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 14,
                "y": 31
              },
              "id": 167,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "gc_rate",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "process_resident_memory_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\" }",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "resident_memory_bytes_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "process_virtual_memory_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "virtual_memory_bytes_{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "process_virtual_memory_max_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "virtual_memory_max_bytes_{{instance}}",
                  "refId": "C"
                }
              ],
              "title": "process_virtual_memory_max",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "decbytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "y": 37
              },
              "id": 110,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "gc_rate",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "process_open_fds{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "open_fds_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "process_max_fds{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "max_fds_{{instance}}",
                  "refId": "B"
                }
              ],
              "title": "open_fds",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 7,
                "y": 37
              },
              "id": 168,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "gc_rate",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "process_open_fds{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "open_fds_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "process_max_fds{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "max_fds_{{instance}}",
                  "refId": "B"
                }
              ],
              "title": "max_fds",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 14,
                "y": 37
              },
              "id": 111,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "gc_rate",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "rate(process_cpu_seconds_total{cluster=\"$cluster\", role=\"$role\"}[1m])",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "cpu_seconds_{{instance}}",
                  "refId": "A"
                }
              ],
              "title": "process_cpu",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "s",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "locale",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "y": 43
              },
              "id": 62,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "go_memstats_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "go_memstats_alloc_bytes_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_total_{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "go_memstats_heap_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_alloc_bytes_{{instance}}",
                  "refId": "C"
                },
                {
                  "expr": "go_memstats_heap_inuse_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_inuse_bytes_{{instance}}",
                  "refId": "D"
                },
                {
                  "expr": "go_memstats_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "sys_bytes_{{instance}}",
                  "refId": "E"
                },
                {
                  "expr": "go_memstats_mallocs_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mallocs_total_{{instance}}",
                  "refId": "F"
                },
                {
                  "expr": "go_memstats_frees_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "frees_total_{{instance}}",
                  "refId": "G"
                },
                {
                  "expr": "go_memstats_other_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "other_sys_bytes_{{instance}}",
                  "refId": "H"
                },
                {
                  "expr": "go_memstats_mcache_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mcache_sys_bytes_{{instance}}",
                  "refId": "I"
                }
              ],
              "title": "alloc_bytes",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "decbytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 7,
                "y": 43
              },
              "id": 172,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "spaceLength": 10,
			  "targets": [
                {
                  "expr": "go_memstats_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "go_memstats_alloc_bytes_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_total_{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "go_memstats_heap_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "heap_alloc_bytes_{{instance}}",
                  "refId": "C"
                },
                {
                  "expr": "go_memstats_heap_inuse_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_inuse_bytes_{{instance}}",
                  "refId": "D"
                },
                {
                  "expr": "go_memstats_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "sys_bytes_{{instance}}",
                  "refId": "E"
                },
                {
                  "expr": "go_memstats_mallocs_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mallocs_total_{{instance}}",
                  "refId": "F"
                },
                {
                  "expr": "go_memstats_frees_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "frees_total_{{instance}}",
                  "refId": "G"
                },
                {
                  "expr": "go_memstats_other_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "other_sys_bytes_{{instance}}",
                  "refId": "H"
                },
                {
                  "expr": "go_memstats_mcache_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mcache_sys_bytes_{{instance}}",
                  "refId": "I"
                }
              ],
              "title": "heap_alloc",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "decbytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 14,
                "y": 43
              },
              "id": 171,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "go_memstats_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "go_memstats_alloc_bytes_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_total_{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "go_memstats_heap_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_alloc_bytes_{{instance}}",
                  "refId": "C"
                },
                {
                  "expr": "go_memstats_heap_inuse_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "heap_inuse_bytes_{{instance}}",
                  "refId": "D"
                },
                {
                  "expr": "go_memstats_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "sys_bytes_{{instance}}",
                  "refId": "E"
                },
                {
                  "expr": "go_memstats_mallocs_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mallocs_total_{{instance}}",
                  "refId": "F"
                },
                {
                  "expr": "go_memstats_frees_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "frees_total_{{instance}}",
                  "refId": "G"
                },
                {
                  "expr": "go_memstats_other_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "other_sys_bytes_{{instance}}",
                  "refId": "H"
                },
                {
                  "expr": "go_memstats_mcache_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mcache_sys_bytes_{{instance}}",
                  "refId": "I"
                }
              ],
              "title": "heap_inuse",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "decbytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "y": 49
              },
              "id": 170,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "go_memstats_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "go_memstats_alloc_bytes_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_total_{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "go_memstats_heap_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_alloc_bytes_{{instance}}",
                  "refId": "C"
                },
                {
                  "expr": "go_memstats_heap_inuse_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_inuse_bytes_{{instance}}",
                  "refId": "D"
                },
                {
                  "expr": "go_memstats_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "sys_bytes_{{instance}}",
                  "refId": "E"
                },
                {
                  "expr": "go_memstats_mallocs_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mallocs_total_{{instance}}",
                  "refId": "F"
                },
                {
                  "expr": "go_memstats_frees_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "frees_total_{{instance}}",
                  "refId": "G"
                },
                {
                  "expr": "go_memstats_other_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "other_sys_bytes_{{instance}}",
                  "refId": "H"
                },
                {
                  "expr": "go_memstats_mcache_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mcache_sys_bytes_{{instance}}",
                  "refId": "I"
                }
              ],
              "title": "sys_bytes",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "decbytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 7,
                "y": 49
              },
              "id": 173,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "go_memstats_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "go_memstats_alloc_bytes_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_total_{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "go_memstats_heap_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_alloc_bytes_{{instance}}",
                  "refId": "C"
                },
                {
                  "expr": "go_memstats_heap_inuse_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_inuse_bytes_{{instance}}",
                  "refId": "D"
                },
                {
                  "expr": "go_memstats_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "sys_bytes_{{instance}}",
                  "refId": "E"
                },
                {
                  "expr": "go_memstats_mallocs_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mallocs_total_{{instance}}",
                  "refId": "F"
                },
                {
                  "expr": "go_memstats_frees_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "frees_total_{{instance}}",
                  "refId": "G"
                },
                {
                  "expr": "go_memstats_other_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "other_sys_bytes_{{instance}}",
				  "refId": "H"
                },
                {
                  "expr": "go_memstats_mcache_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mcache_sys_bytes_{{instance}}",
                  "refId": "I"
                }
              ],
              "title": "other_sys",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "decbytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ]
            },
            {
              "dashLength": 10,
              "fill": 1,
              "gridPos": {
                "h": 6,
                "w": 7,
                "x": 14,
                "y": 49
              },
              "id": 174,
              "legend": {
                "show": true
              },
              "lines": true,
              "linewidth": 1,
              "nullPointMode": "null",
              "pointradius": 5,
              "renderer": "flot",
              "spaceLength": 10,
              "targets": [
                {
                  "expr": "go_memstats_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_{{instance}}",
                  "refId": "A"
                },
                {
                  "expr": "go_memstats_alloc_bytes_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "alloc_bytes_total_{{instance}}",
                  "refId": "B"
                },
                {
                  "expr": "go_memstats_heap_alloc_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_alloc_bytes_{{instance}}",
                  "refId": "C"
                },
                {
                  "expr": "go_memstats_heap_inuse_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "heap_inuse_bytes_{{instance}}",
                  "refId": "D"
                },
                {
                  "expr": "go_memstats_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "sys_bytes_{{instance}}",
                  "refId": "E"
                },
                {
                  "expr": "go_memstats_mallocs_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "mallocs_total_{{instance}}",
                  "refId": "F"
                },
                {
                  "expr": "go_memstats_frees_total{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "frees_total_{{instance}}",
                  "refId": "G"
                },
                {
                  "expr": "go_memstats_other_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "hide": true,
                  "intervalFactor": 1,
                  "legendFormat": "other_sys_bytes_{{instance}}",
                  "refId": "H"
                },
                {
                  "expr": "go_memstats_mcache_sys_bytes{app=\"$app\", cluster=\"$cluster\", role=\"$role\"}",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "mcache_sys_bytes_{{instance}}",
                  "refId": "I"
                }
              ],
              "title": "mcache_sys",
              "tooltip": {
                "shared": true,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "mode": "time",
                "show": true
              },
              "yaxes": [
                {
                  "format": "decbytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ]
            }
          ],
          "title": "GoRuntime",
          "type": "row"
        }
      ],
      "schemaVersion": 20,
      "style": "dark",
      "tags": [],
      "templating": {
        "list": [
          {
            "allValue": null,
            "current": {
              "selected": true,
              "text": "cfs",
              "value": "cfs"
            },
            "hide": 2,
            "includeAll": false,
            "label": "App",
            "multi": false,
            "name": "app",
            "options": [
              {
                "selected": true,
                "text": "cfs",
                "value": "cfs"
              }
            ],
            "query": "cfs",
            "skipUrlSync": false,
            "type": "custom"
          },
          {
            "allValue": null,
            "current": {
              "text": "chubaofs01",
              "value": "chubaofs01"
            },
            "datasource": "Prometheus",
            "definition": "label_values(up{app=\"$app\"}, cluster)",
            "hide": 0,
            "includeAll": false,
            "label": "Cluster",
            "multi": false,
            "name": "cluster",
            "options": [],
            "query": "label_values(up{app=\"$app\"}, cluster)",
            "refresh": 1,
            "regex": "",
            "skipUrlSync": false,
            "sort": 0,
            "tagValuesQuery": "",
            "tags": [],
            "tagsQuery": "",
            "type": "query",
            "useTags": false
          },
          {
            "allValue": null,
            "current": {
              "text": "master",
              "value": "master"
            },
            "datasource": "Prometheus",
            "definition": "label_values(up{app=\"$app\", cluster=\"$cluster\"}, role)",
            "hide": 0,
            "includeAll": false,
            "label": "Role",
            "multi": false,
            "name": "role",
            "options": [],
            "query": "label_values(up{app=\"$app\", cluster=\"$cluster\"}, role)",
            "refresh": 1,
            "regex": "",
            "skipUrlSync": false,
            "sort": 0,
            "tagValuesQuery": "",
            "tags": [],
            "tagsQuery": "",
            "type": "query",
            "useTags": false
          },
          {
            "allValue": null,
            "current": {
              "text": "5",
              "value": "5"
            },
            "hide": 0,
            "includeAll": false,
            "label": "topK",
            "multi": false,
            "name": "topk",
            "options": [
              {
                "selected": true,
                "text": "5",
                "value": "5"
              },
              {
                "text": "10",
                "value": "10"
              },
              {
                "text": "20",
                "value": "20"
              },
              {
                "text": "30",
                "value": "30"
              },
              {
                "text": "50",
                "value": "50"
              },
              {
                "text": "100",
                "value": "100"
              }
            ],
            "query": "5,10,20,30,50,100",
            "skipUrlSync": false,
            "type": "custom"
          },
          {
            "allValue": null,
            "current": {
              "tags": [],
              "text": "OpMetaInodeGet",
              "value": [
                "OpMetaInodeGet"
              ]
            },
            "hide": 0,
            "includeAll": true,
            "label": "MetaNodeOp",
            "multi": true,
            "name": "mnOp",
            "options": [
              {
                "selected": false,
                "text": "All",
                "value": "$__all"
              },
              {
                "selected": false,
                "text": "OpMetaCreateInode",
                "value": "OpMetaCreateInode"
              },
              {
                "selected": false,
                "text": "OpMetaLinkInode",
                "value": "OpMetaLinkInode"
              },
              {
                "selected": false,
                "text": "OpMetaFreeInodesOnRaftFollower",
                "value": "OpMetaFreeInodesOnRaftFollower"
              },
              {
                "selected": false,
                "text": "OpMetaUnlinkInode",
                "value": "OpMetaUnlinkInode"
              },
              {
                "selected": true,
                "text": "OpMetaInodeGet",
                "value": "OpMetaInodeGet"
              },
              {
                "selected": false,
                "text": "OpMetaEvictInode",
                "value": "OpMetaEvictInode"
              },
              {
                "selected": false,
                "text": "OpMetaSetattr",
                "value": "OpMetaSetattr"
              },
              {
                "selected": false,
                "text": "OpMetaCreateDentry",
                "value": "OpMetaCreateDentry"
              },
              {
                "selected": false,
                "text": "OpMetaDeleteDentry",
                "value": "OpMetaDeleteDentry"
              },
              {
                "selected": false,
                "text": "OpMetaUpdateDentry",
                "value": "OpMetaUpdateDentry"
              },
              {
                "selected": false,
                "text": "OpMetaReadDir",
                "value": "OpMetaReadDir"
              },
              {
                "selected": false,
                "text": "OpCreateMetaPartition",
                "value": "OpCreateMetaPartition"
              },
              {
                "selected": false,
                "text": "OpMetaNodeHeartbeat",
                "value": "OpMetaNodeHeartbeat"
              },
              {
                "selected": false,
                "text": "OpMetaExtentsAdd",
                "value": "OpMetaExtentsAdd"
              },
              {
                "selected": false,
                "text": "OpMetaExtentsList",
                "value": "OpMetaExtentsList"
              },
              {
                "selected": false,
                "text": "OpMetaExtentsDel",
                "value": "OpMetaExtentsDel"
              },
              {
                "selected": false,
                "text": "OpMetaTruncate",
                "value": "OpMetaTruncate"
              },
              {
                "selected": false,
                "text": "OpMetaLookup",
                "value": "OpMetaLookup"
              },
              {
                "selected": false,
                "text": "OpDeleteMetaPartition",
                "value": "OpDeleteMetaPartition"
              },
              {
                "selected": false,
                "text": "OpUpdateMetaPartition",
                "value": "OpUpdateMetaPartition"
              },
              {
                "selected": false,
                "text": "OpLoadMetaPartition",
                "value": "OpLoadMetaPartition"
              },
              {
                "selected": false,
                "text": "OpDecommissionMetaPartition",
                "value": "OpDecommissionMetaPartition"
              },
              {
                "selected": false,
                "text": "OpAddMetaPartitionRaftMember",
                "value": "OpAddMetaPartitionRaftMember"
              },
              {
                "selected": false,
                "text": "OpRemoveMetaPartitionRaftMember",
                "value": "OpRemoveMetaPartitionRaftMember"
              },
              {
                "selected": false,
                "text": "OpMetaPartitionTryToLeader",
                "value": "OpMetaPartitionTryToLeader"
              },
              {
                "selected": false,
                "text": "OpMetaBatchInodeGet",
                "value": "OpMetaBatchInodeGet"
              },
              {
                "selected": false,
                "text": "OpMetaDeleteInode",
                "value": "OpMetaDeleteInode"
              },
              {
                "selected": false,
                "text": "OpMetaBatchExtentsAdd",
                "value": "OpMetaBatchExtentsAdd"
              },
              {
                "selected": false,
                "text": "OpMetaSetXAttr",
                "value": "OpMetaSetXAttr"
              },
              {
                "selected": false,
                "text": "OpMetaGetXAttr",
                "value": "OpMetaGetXAttr"
              },
              {
                "selected": false,
                "text": "OpMetaBatchGetXAttr",
                "value": "OpMetaBatchGetXAttr"
              },
              {
                "selected": false,
                "text": "OpMetaRemoveXAttr",
                "value": "OpMetaRemoveXAttr"
              },
              {
                "selected": false,
                "text": "OpMetaListXAttr",
                "value": "OpMetaListXAttr"
              },
              {
                "selected": false,
                "text": "OpCreateMultipart",
                "value": "OpCreateMultipart"
              },
              {
                "selected": false,
                "text": "OpListMultiparts",
                "value": "OpListMultiparts"
              },
              {
                "selected": false,
                "text": "OpRemoveMultipart",
                "value": "OpRemoveMultipart"
              },
              {
                "selected": false,
                "text": "OpAddMultipartPart",
                "value": "OpAddMultipartPart"
              },
              {
                "selected": false,
                "text": "OpGetMultipart",
                "value": "OpGetMultipart"
              }
            ],
            "query": "OpMetaCreateInode, OpMetaLinkInode, OpMetaFreeInodesOnRaftFollower, OpMetaUnlinkInode, OpMetaInodeGet, OpMetaEvictInode, OpMetaSetattr, OpMetaCreateDentry, OpMetaDeleteDentry, OpMetaUpdateDentry, OpMetaReadDir, OpCreateMetaPartition, OpMetaNodeHeartbeat, OpMetaExtentsAdd, OpMetaExtentsList, OpMetaExtentsDel, OpMetaTruncate, OpMetaLookup, OpDeleteMetaPartition, OpUpdateMetaPartition, OpLoadMetaPartition, OpDecommissionMetaPartition, OpAddMetaPartitionRaftMember, OpRemoveMetaPartitionRaftMember, OpMetaPartitionTryToLeader, OpMetaBatchInodeGet, OpMetaDeleteInode, OpMetaBatchExtentsAdd, OpMetaSetXAttr, OpMetaGetXAttr, OpMetaBatchGetXAttr, OpMetaRemoveXAttr, OpMetaListXAttr, OpCreateMultipart, OpListMultiparts, OpRemoveMultipart, OpAddMultipartPart, OpGetMultipart",
            "skipUrlSync": false,
            "type": "custom"
          },
          {
            "allValue": null,
            "current": {
              "tags": [],
              "text": "OpCreateExtent",
              "value": [
                "OpCreateExtent"
              ]
            },
            "hide": 0,
            "includeAll": true,
            "label": "DataNodeOp",
            "multi": true,
            "name": "dnOp",
            "options": [
              {
                "selected": false,
                "text": "All",
                "value": "$__all"
              },
              {
                "selected": true,
                "text": "OpCreateExtent",
                "value": "OpCreateExtent"
              },
              {
                "selected": false,
                "text": "OpWrite",
                "value": "OpWrite"
              },
              {
                "selected": false,
                "text": "OpSyncWrite",
                "value": "OpSyncWrite"
              },
              {
                "selected": false,
                "text": "OpStreamRead",
                "value": "OpStreamRead"
              },
              {
                "selected": false,
                "text": "OpStreamFollowerRead",
                "value": "OpStreamFollowerRead"
              },
              {
                "selected": false,
                "text": "OpExtentRepairRead",
                "value": "OpExtentRepairRead"
              },
              {
                "selected": false,
                "text": "OpMarkDelete",
                "value": "OpMarkDelete"
              },
              {
                "selected": false,
                "text": "OpRandowWrite",
                "value": "OpRandowWrite"
              },
              {
                "selected": false,
                "text": "OpSyncRandomWrite",
                "value": "OpSyncRandomWrite"
              },
              {
                "selected": false,
                "text": "OpNotifyReplicasToRepair",
                "value": "OpNotifyReplicasToRepair"
              },
              {
                "selected": false,
                "text": "OpGetAllWatermarks",
                "value": "OpGetAllWatermarks"
              },
              {
                "selected": false,
                "text": "OpCreateDataPartition",
                "value": "OpCreateDataPartition"
              },
              {
                "selected": false,
                "text": "OpLoadDataPratition",
                "value": "OpLoadDataPratition"
              },
              {
                "selected": false,
                "text": "OpDeleteDataPartition",
                "value": "OpDeleteDataPartition"
              },
              {
                "selected": false,
                "text": "OpDataNodeHeartbeat",
                "value": "OpDataNodeHeartbeat"
              },
              {
                "selected": false,
                "text": "OpGetAppliedId",
                "value": "OpGetAppliedId"
              },
              {
                "selected": false,
                "text": "OpDecommissionDataPartition",
                "value": "OpDecommissionDataPartition"
              },
              {
                "selected": false,
                "text": "OpAddDataPartitionRaftMember",
				"value": "OpAddDataPartitionRaftMember"
              },
              {
                "selected": false,
                "text": "OpDataPartitionTryToLeader",
                "value": "OpDataPartitionTryToLeader"
              },
              {
                "selected": false,
                "text": "OpGetPartitionSize",
                "value": "OpGetPartitionSize"
              },
              {
                "selected": false,
                "text": "OpGetMaxExtentIDAndPartitionSize",
                "value": "OpGetMaxExtentIDAndPartitionSize"
              },
              {
                "selected": false,
                "text": "OpReadTinyDeleteRecord",
                "value": "OpReadTinyDeleteRecord"
              },
              {
                "selected": false,
                "text": "OpBroadcastMinAppliedID",
                "value": "OpBroadcastMinAppliedID"
              }
            ],
            "query": "OpCreateExtent,OpWrite,OpSyncWrite,OpStreamRead,OpStreamFollowerRead,OpExtentRepairRead,OpMarkDelete,OpRandowWrite,OpSyncRandomWrite,OpNotifyReplicasToRepair,OpGetAllWatermarks,OpCreateDataPartition,OpLoadDataPratition,OpDeleteDataPartition,OpDataNodeHeartbeat,OpGetAppliedId,OpDecommissionDataPartition,OpAddDataPartitionRaftMember,OpDataPartitionTryToLeader,OpGetPartitionSize,OpGetMaxExtentIDAndPartitionSize,OpReadTinyDeleteRecord,OpBroadcastMinAppliedID",
            "skipUrlSync": false,
            "type": "custom"
          },
          {
            "allValue": null,
            "current": {
              "text": "mkdir",
              "value": [
                "mkdir"
              ]
            },
            "hide": 0,
            "includeAll": true,
            "label": "fuseOp",
            "multi": true,
            "name": "fuseOp",
            "options": [
              {
                "text": "All",
                "value": "$__all"
              },
              {
                "selected": true,
                "text": "mkdir",
                "value": "mkdir"
              },
              {
                "text": "mknod",
                "value": "mknod"
              },
              {
                "text": "filecreate",
                "value": "filecreate"
              },
              {
                "text": "link",
                "value": "link"
              },
              {
                "text": "symlink",
                "value": "symlink"
              },
              {
                "text": "rmdir",
                "value": "rmdir"
              },
              {
                "text": "unlink",
                "value": "unlink"
              },
              {
                "text": "readdir",
                "value": "readdir"
              },
              {
                "text": "rename",
                "value": "rename"
              },
              {
                "text": "fileread",
                "value": "fileread"
              },
              {
                "text": "filewrite",
                "value": "filewrite"
              },
              {
                "text": "filesync",
                "value": "filesync"
              }
            ],
            "query": "mkdir,mknod,filecreate,link,symlink,rmdir,unlink,readdir,rename,fileread,filewrite,filesync",
            "skipUrlSync": false,
            "type": "custom"
          },
          {
            "allValue": null,
            "current": {
              "tags": [],
              "text": "ListObjects",
              "value": [
                "ListObjects"
              ]
            },
            "hide": 0,
            "includeAll": true,
            "label": "objectNodeOp",
            "multi": true,
            "name": "obOp",
            "options": [
              {
                "text": "All",
                "value": "$__all"
              },
              {
                "text": "GetObject",
                "value": "GetObject"
              },
              {
                "text": "PutObject",
                "value": "PutObject"
              },
              {
                "text": "CopyObject",
                "value": "CopyObject"
              },
              {
                "selected": true,
                "text": "ListObjects",
                "value": "ListObjects"
              },
              {
                "text": "DeleteObject",
                "value": "DeleteObject"
              },
              {
                "text": "DeleteObjects",
                "value": "DeleteObjects"
              },
              {
                "text": "HeadObject",
                "value": "HeadObject"
              },
              {
                "text": "CreateBucket",
                "value": "CreateBucket"
              },
              {
                "text": "DeleteBucket",
                "value": "DeleteBucket"
              },
              {
                "text": "HeadBucket",
                "value": "HeadBucket"
              },
              {
                "text": "ListBuckets",
                "value": "ListBuckets"
              },
              {
                "text": "GetBucketPolicy",
                "value": "GetBucketPolicy"
              },
              {
                "text": "PutBucketPolicy",
                "value": "PutBucketPolicy"
              },
              {
                "text": "DeleteBucketPolicy",
                "value": "DeleteBucketPolicy"
              },
              {
                "text": "GetBucketAcl",
                "value": "GetBucketAcl"
              },
              {
                "text": "PutBucketAcl",
                "value": "PutBucketAcl"
              },
              {
                "text": "GetObjectAcl",
                "value": "GetObjectAcl"
              },
              {
                "text": "PutObjectAcl",
                "value": "PutObjectAcl"
              },
              {
                "text": "CreateMultipartUpload",
                "value": "CreateMultipartUpload"
              },
              {
                "text": "ListMultipartUploads",
                "value": "ListMultipartUploads"
              },
              {
                "text": "UploadPart",
                "value": "UploadPart"
              },
              {
                "text": "ListParts",
                "value": "ListParts"
              },
              {
                "text": "CompleteMultipartUpload",
                "value": "CompleteMultipartUpload"
              },
              {
                "text": "AbortMultipartUpload",
                "value": "AbortMultipartUpload"
              },
              {
                "text": "GetBucketLocation",
                "value": "GetBucketLocation"
              },
              {
                "text": "GetObjectXAttr",
                "value": "GetObjectXAttr"
              },
              {
                "text": "PutObjectXAttr",
                "value": "PutObjectXAttr"
              },
              {
                "text": "ListObjectXAttrs",
                "value": "ListObjectXAttrs"
              },
              {
                "text": "DeleteObjectXAttr",
                "value": "DeleteObjectXAttr"
              },
              {
                "text": "GetObjectTagging",
                "value": "GetObjectTagging"
              },
              {
                "text": "PutObjectTagging",
                "value": "PutObjectTagging"
              },
              {
                "text": "DeleteObjectTagging",
                "value": "DeleteObjectTagging"
              },
              {
                "text": "GetBucketTagging",
                "value": "GetBucketTagging"
              },
              {
                "text": "PutBucketTagging",
                "value": "PutBucketTagging"
              },
              {
                "text": "DeleteBucketTagging",
                "value": "DeleteBucketTagging"
              },
              {
                "text": "Read",
                "value": "Read"
              },
              {
                "text": "Write",
                "value": "Write"
              }
            ],
            "query": "GetObject, PutObject, CopyObject, ListObjects, DeleteObject, DeleteObjects, HeadObject, CreateBucket, DeleteBucket, HeadBucket, ListBuckets, GetBucketPolicy, PutBucketPolicy, DeleteBucketPolicy, GetBucketAcl, PutBucketAcl, GetObjectAcl, PutObjectAcl, CreateMultipartUpload, ListMultipartUploads, UploadPart, ListParts, CompleteMultipartUpload, AbortMultipartUpload, GetBucketLocation, GetObjectXAttr, PutObjectXAttr, ListObjectXAttrs, DeleteObjectXAttr, GetObjectTagging, PutObjectTagging, DeleteObjectTagging, GetBucketTagging, PutBucketTagging, DeleteBucketTagging, Read, Write,",
            "skipUrlSync": false,
            "type": "custom"
          },
          {
            "allValue": null,
            "current": {
              "text": "ltptest",
              "value": "ltptest"
            },
            "datasource": "Prometheus",
            "definition": "label_values(cfs_master_vol_total_GB{app=\"$app\",cluster=\"$cluster\"}, volName)",
            "hide": 0,
            "includeAll": false,
            "label": "Vol",
            "multi": false,
            "name": "vol",
            "options": [],
            "query": "label_values(cfs_master_vol_total_GB{app=\"$app\",cluster=\"$cluster\"}, volName)",
            "refresh": 1,
            "regex": "",
            "skipUrlSync": false,
            "sort": 0,
            "tagValuesQuery": "",
            "tags": [],
            "tagsQuery": "",
            "type": "query",
            "useTags": false
          },
          {
            "allValue": null,
            "current": {
              "text": "192.168.0.11:17010",
              "value": "192.168.0.11:17010"
            },
            "datasource": "Prometheus",
            "definition": "label_values(up{app=\"$app\", role=\"$role\", cluster=\"$cluster\"}, instance)",
            "hide": 0,
            "includeAll": false,
            "label": "Instance",
            "multi": false,
            "name": "instance",
            "options": [],
            "query": "label_values(up{app=\"$app\", role=\"$role\", cluster=\"$cluster\"}, instance)",
            "refresh": 1,
            "regex": "",
            "skipUrlSync": false,
            "sort": 0,
            "tagValuesQuery": "",
            "tags": [],
            "tagsQuery": "",
            "type": "query",
            "useTags": false
          },
          {
            "allValue": null,
            "current": {
              "text": "192.168.0.11",
              "value": "192.168.0.11"
            },
            "datasource": "Prometheus",
            "definition": "label_values(go_info{instance=\"$instance\", cluster=\"$cluster\"}, instance)",
            "hide": 2,
            "includeAll": false,
            "label": "Host",
            "multi": false,
            "name": "hostip",
            "options": [],
            "query": "label_values(go_info{instance=\"$instance\", cluster=\"$cluster\"}, instance)",
            "refresh": 1,
            "regex": "/([^:]+):.*/",
            "skipUrlSync": false,
            "sort": 0,
            "tagValuesQuery": "",
            "tags": [],
            "tagsQuery": "",
            "type": "query",
            "useTags": false
          },
          {
            "allValue": null,
            "current": {
              "text": "192.168.0.11:17010",
              "value": "192.168.0.11:17010"
            },
            "datasource": "Prometheus",
            "definition": "label_values(up{app=\"$app\", role=\"master\", cluster=\"$cluster\"}, instance)",
            "hide": 2,
            "includeAll": false,
            "label": "MasterInstance",
            "multi": false,
            "name": "master_instance",
            "options": [],
            "query": "label_values(up{app=\"$app\", role=\"master\", cluster=\"$cluster\"}, instance)",
            "refresh": 1,
            "regex": "",
            "skipUrlSync": false,
            "sort": 0,
            "tagValuesQuery": "",
            "tags": [],
            "tagsQuery": "",
            "type": "query",
            "useTags": false
          }
        ]
      },
      "time": {
        "from": "now-15m",
        "to": "now"
      },
      "timepicker": {
        "refresh_intervals": [
          "5s",
          "10s",
          "30s",
          "1m",
          "5m",
          "15m",
          "30m",
          "1h",
          "2h",
          "1d"
        ],
        "time_options": [
          "5m",
          "15m",
          "1h",
          "6h",
          "12h",
          "24h",
          "2d",
          "7d",
          "30d"
        ]
      },
      "timezone": "",
	  "title": "ChubaoFS",
      "uid": "J8XJyOmZk1",
      "version": 1
    }`,
		},
	}

	return cfg
}
