package plugins

import (
	// registry all plugins
	_ "bes-agent/collector/plugins/apache"
	_ "bes-agent/collector/plugins/docker"
	_ "bes-agent/collector/plugins/haproxy"
	_ "bes-agent/collector/plugins/memcached"
	_ "bes-agent/collector/plugins/mongodb"
	_ "bes-agent/collector/plugins/mysql"
	_ "bes-agent/collector/plugins/nginx"
	_ "bes-agent/collector/plugins/phpfpm"
	_ "bes-agent/collector/plugins/postgres"
	_ "bes-agent/collector/plugins/redis"
	_ "bes-agent/collector/plugins/system"
)
