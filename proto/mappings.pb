# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# proto-file: proto/mappings.proto
# proto-message: Mappings

nodes {
  subpath {path: "/components/component[name=name_value]"}

  map {key: "name_value", value: "index"}

  children {
    subpath: {path: "name"}
    bind: "cpu_name"
  }

  children {
    subpath {path: "cpu/utilization/state/avg"}
    bind: "avg_cpu_util"
  }
}

nodes {
  subpath {path: "/system"}

  children {
    subpath {path: "state/boot-time"}
    bind: "boot_time"
  }

  children {
    subpath {path: "memory"}

    children {
      subpath: {path: "state"}

      children {
        subpath: {
          # Total physical memory in bytes.
          path: "physical"
          revisions: "0.4.1"
        }
        bind: "total_memory_B"
      }

      children {
        subpath: {
          # Used memory in bytes.
          path: "reserved"
          revisions: "0.4.1"
        }
        bind: "used_memory"
      }
    }
  }
}

nodes {
  subpath { path: "/interfaces/interface[name=name_value]" }
  # NB: There does not appear to be a way to retrieve interface names via SNMP, so use indexes.
  map {key: "name_value", value: "interface_index"}

  children {
    subpath: {path: "state"}

    children {
      subpath: {path: "ifindex"}
      bind: "interface_index"
    }

    children {
      subpath: {path: "admin-status"}
      bind: "admin_status"
    }

    children {
      # The time at which the last system change occurred, in ns, relative to the Unix Epoch.
      subpath: {path: "last-change"}
      bind: "last_change_absolute"


    }

    children {
      subpath: {path: "counters"}

      children {
        subpath {path: "in-broadcast-pkts"}
        bind: "in_broadcast_packets"
      }

      children {
        subpath {path: "in-multicast-pkts"}
        bind: "in_multicast_packets"
      }

      children {
        subpath {path: "in-unicast-pkts"}
        bind: "in_unicast_packets"
      }

      children {
        subpath {path: "in-octets"}
        bind: "in_octets"
      }

      children {
        subpath {path: "out-broadcast-pkts"}
        bind: "out_broadcast_packets"
      }

      children {
        subpath {path: "out-multicast-pkts"}
        bind: "out_multicast_packets"
      }

      children {
        subpath {path: "out-unicast-pkts"}
        bind: "out_unicast_packets"
      }

      children {
        subpath {path: "out-octets"}
        bind: "out_octets"
      }

      children {
        subpath {path: "in-discards"}
        bind: "in_discards"
      }

      children {
        subpath {path: "in-errors"}
        bind: "in_errors"
      }
    }
  }
}