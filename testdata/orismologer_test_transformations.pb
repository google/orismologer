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
# proto-message: Transformations

transformations {
  bind: "cpu_name"
  expressions: "cpu_name_aruba_oid"

  noc_paths {
    bind: "cpu_name_aruba_oid"
    oids: "1.3.6.1.4.1.14823.2.2.1.1.1.9.1.2.index"
    samples: "Network Processor CPU10"
  }
}

######################################

transformations {
  bind: "boot_time"
  expressions: "time_since_epoch(system_time_aruba, '2006-01-02 15:04:05', 's') - system_up_time"
  expressions: "time_since_epoch(system_time_cisco, 'ntp', 's') - system_up_time"

  noc_paths {
    bind: "system_time_aruba"
    oids: "1.3.6.1.4.1.14823.2.2.1.2.1.6"
    samples: "2018-12-18 15:15:59"
  }
  noc_paths {
    bind: "system_time_cisco"
    oids: "1.3.6.1.4.1.9.9.168.1.1.10"
    samples: "dfc4 0b68 8147 af78"
  }
}

transformations: {
  bind: "last_change_absolute"
  expressions: "(boot_time + to_int(last_change_relative)) * 1000"

  noc_paths {
    bind: "last_change_relative"
    oids: "1.3.6.1.2.1.2.2.1.9.interface_index"
    samples: "50"
  }
}

transformations {
  bind: "system_up_time"
  expressions: "to_int(system_up_time_100) / 100"

  noc_paths {
    bind: "system_up_time_100"
    oids: "1.3.6.1.2.1.1.3"
    samples: "2000000000"
  }
}

######################################

transformations {
  bind: "used_memory"
  expressions: "used_memory_cisco"
  expressions: "to_int(used_memory_KB_oid_aruba) * 1000"

  noc_paths {
    bind: "used_memory_KB_oid_aruba"
    oids: "1.3.6.1.4.1.14823.2.2.1.1.1.11.1.3"
    samples: "2179200"
  }
}

transformations {
  bind: "total_memory_B"
  expressions: "to_int(total_memory_aruba) * 1000"
  expressions: "int(free_memory_cisco) + used_memory_cisco"

  noc_paths {
    bind: "total_memory_aruba"
    oids: "1.3.6.1.4.1.14823.2.2.1.1.1.11.1.2"
    samples: "5172096"
  }
  noc_paths {
    bind: "free_memory_cisco"
    oids: "1.3.6.1.4.1.9.9.48.1.1.1.6.1"
    samples: "556513160"
  }
}

transformations {
  bind: "used_memory_cisco"
  expressions: "to_int(used_memory_B_cisco)"

  noc_paths {
    bind: "used_memory_b_cisco"
    oids: "1.3.6.1.4.1.9.9.48.1.1.1.5.1"
    samples: "383014872"
  }
}
