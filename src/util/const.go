package util

import (
)

const (
    POODLE_EPOCH_MILLIS             = 30 * 1000

    DEFAULT_DRIFT_MILLIS_LOW        = 300
    DEFAULT_DRIFT_MILLIS_HIGH       = 500

    DEFAULT_ETC_DIR                 = "/etc/poodle"
    DEFAULT_LIB_DIR                 = "/var/lib/poodle"
    DEFAULT_LOG_DIR                 = "/var/log/poodle"

    DEFAULT_SECRET                  = "poodle"

    DEFAULT_UDP_PORT                = 31415
    DEFAULT_QUIC_PORT               = 31416

    CLS_NODE                        = 1
    CLS_CLUSTER                     = 2
    CLS_UNIVERSE                    = 3
    CLS_SERVICE                     = 4
    CLS_FEDERATION                  = 5
)
