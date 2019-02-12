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
  bind: "used_memory_cisco"
  expressions: "to_int(used_memory_B_cisco)"

  noc_paths {
    bind: "used_memory_b_cisco"
    # Used memory in B.
    # Index .1 corresponds to "processor" memory (processes). .2 is "I/O" (packet queues etc.)
    oids: "1.3.6.1.4.1.9.9.48.1.1.1.5.1"  # Cisco: {iso(1) identified-organization(3) dod(6) internet(1) private(4) enterprise(1) 9 ciscoMgmt(9) ciscoMemoryPoolMIB(48) ciscoMemoryPoolObjects(1) ciscoMemoryPoolTable(1) ciscoMemoryPoolEntry(1) ciscoMemoryPoolUsed(5)}
    samples: "383014872"
  }
}

transformations {
  # Time at which the system booted, in seconds, relative to the Linux Epoch.
  bind: "boot_time"
  expressions: "time_since_epoch(system_time_aruba, '2006-01-02 15:04:05', 's') - system_up_time"
  expressions: "time_since_epoch(system_time_cisco, 'ntp', 's') - system_up_time"

  noc_paths {
    bind: "system_time_aruba"
    oids: "1.3.6.1.4.1.14823.2.2.1.2.1.6"  # Aruba: {iso(1) identified-organization(3) dod(6) internet(1) private(4) enterprise(1) 14823 arubaEnterpriseMibModules(2) switch(2) wlsxEnterpriseMibModules(1) wlsxSystemExtMIB(2) wlsxSystemExtGroup(1) wlsxSysExtSwitchDate(6)}
    samples: "2018-12-18 15:15:59"
  }
  noc_paths {
    bind: "system_time_cisco"
    oids: "1.3.6.1.4.1.9.9.168.1.1.10"  # Cisco.
    samples: "dfc4 0b68 8147 af78"  # NTP timestamp (64bit, first 32b = seconds since 1900-01-01, second 32b = fractional component)
  }
}

transformations {
  # The number of seconds since the system booted.
  bind: "system_up_time"
  expressions: "to_int(system_up_time_100) / 100"

  noc_paths {
    # Time since system booted in 100ths of a second.
    bind: "system_up_time_100"
    oids: "1.3.6.1.2.1.1.3"  # Standard MIB: {iso(1) identified-organization(3) dod(6) internet(1) mgmt(2) mib-2(1) system(1) sysUpTime(3)}
    samples: "2026708237"
  }
}

transformations {
  bind: "cpu_name"
  expressions: "cpu_name_aruba_oid"

  noc_paths {
    bind: "cpu_name_aruba_oid"
    oids: "1.3.6.1.4.1.14823.2.2.1.1.1.9.1.2.index"  # Aruba: {iso(1) identified-organization(3) dod(6) internet(1) private(4) enterprise(1) 14823 arubaEnterpriseMibModules(2) switch(2) wlsxEnterpriseMibModules(1) wlsxSwitchMIB(1) wlsxSystemXGroup(1) wlsxSysXProcessorTable(9) wlsxSysXProcessorEntry(1) sysXProcessorDescr(2)}
    samples: "Network Processor CPU10"
  }
}

transformations {
  bind: "avg_cpu_util"
  expressions: "to_int(avg_cpu_util_oid)"

  noc_paths {
    bind: "avg_cpu_util_oid"
    # CPU utilisation as a percentage, averaged over the past minute.
    oids: "1.3.6.1.4.1.14823.2.2.1.1.1.9.1.3.index"  # Aruba: {iso(1) identified-organization(3) dod(6) internet(1) private(4) enterprise(1) 14823 arubaEnterpriseMibModules(2) switch(2) wlsxEnterpriseMibModules(1) wlsxSwitchMIB(1) wlsxSystemXGroup(1) wlsxSysXProcessorTable(9) wlsxSysXProcessorEntry(1) sysXProcessorLoad(3)}
    oids: "1.3.6.1.4.1.9.9.109.1.1.1.1.7.index"  # Cisco: {iso(1) identified-organization(3) dod(6) internet(1) private(4) enterprise(1) 9 ciscoMgmt(9) ciscoProcessMIB(109) ciscoProcessMIBObjects(1) cpmCPU(1) cpmCPUTotalTable(1) cpmCPUTotalEntry(1) cpmCPUTotal1minRev(7)}
    oids: "1.3.6.1.4.1.2636.3.1.13.1.8.index"  # Juniper TODO: Cannot find good reference to confirm that this is the right OID.
    oids: "1.3.6.1.2.1.25.3.3.1.2.index"  # Standard MIB: {iso(1) identified-organization(3) dod(6) internet(1) mgmt(2) mib-2(1) host(25) hrDevice(3) hrProcessorTable(3) hrProcessorEntry(1) hrProcessorLoad(2)}
    samples: "6"
    samples: "19"
    samples: "0"
    samples: "100"
  }
}

transformations {
  bind: "total_memory_B"
  expressions: "to_int(total_memory_aruba) * 1000"
  expressions: "int(free_memory_cisco) + used_memory_cisco"

  noc_paths {
    bind: "total_memory_aruba"
    # Total memory in KB (Aruba).
    oids: "1.3.6.1.4.1.14823.2.2.1.1.1.11.1.2"  # {iso(1) identified-organization(3) dod(6) internet(1) private(4) enterprise(1) 14823 arubaEnterpriseMibModules(2) switch(2) wlsxEnterpriseMibModules(1) wlsxSwitchMIB(1) wlsxSystemXGroup(1) wlsxSysXMemoryTable(11) wlsxSysXMemoryEntry(1) sysXMemorySize(2)}
    samples: "5172096"
  }
  noc_paths {
    bind: "free_memory_cisco"
    # Free memory in bytes (Cisco).
    # Index .1 corresponds to "processor" memory (processes). .2 is "I/O" (packet queues etc.)
    oids: "1.3.6.1.4.1.9.9.48.1.1.1.6.1"  # {iso(1) identified-organization(3) dod(6) internet(1) private(4) enterprise(1) 9 ciscoMgmt(9) ciscoMemoryPoolMIB(48) ciscoMemoryPoolObjects(1) ciscoMemoryPoolTable(1) ciscoMemoryPoolEntry(1) ciscoMemoryPoolFree(6)}
    samples: "556513160"
  }
}

transformations {
  bind: "used_memory"
  expressions: "used_memory_cisco"
  expressions: "to_int(used_memory_KB_oid_aruba) * 1000"

  noc_paths {
    # Used memory in KB.
    bind: "used_memory_KB_oid_aruba"
    oids: "1.3.6.1.4.1.14823.2.2.1.1.1.11.1.3"  # {iso(1) identified-organization(3) dod(6) internet(1) private(4) enterprise(1) 14823 arubaEnterpriseMibModules(2) switch(2) wlsxEnterpriseMibModules(1) wlsxSwitchMIB(1) wlsxSystemXGroup(1) wlsxSysXMemoryTable(11) wlsxSysXMemoryEntry(1) sysXMemoryUsed(3)}
    samples: "2179200"
  }
}

transformations {
  bind: "interface_index"
  expressions: "to_int(interface_index_raw)"

  noc_paths {
    bind: "interface_index_raw"
    oids: "1.3.6.1.2.1.2.2.1.1.interface_index"
    samples: "1"
    samples: "134217728"
  }
}

transformations {
  bind: "admin_status"
  expressions: "to_int(admin_status_raw)"

  noc_paths {
    bind: "admin_status_raw"
    oids: "1.3.6.1.2.1.2.2.1.7.interface_index"  # Standard MIB.
    samples: "1"  # Up.
    samples: "2"  # Down.
    samples: "3"  # Testing.
  }
}

transformations: {
  bind: "last_change_absolute"  # Relative to Unix epoch.
  expressions: "(boot_time + to_int(last_change_relative)) * 1000000000"

  noc_paths {
    bind: "last_change_relative"  # Seconds since system booted, when last change occurred.
    oids: "1.3.6.1.2.1.2.2.1.9.interface_index"  # Standard MIB.
    samples: "1557122348"
    samples: "9624"
  }
}

transformations {
  bind: "in_broadcast_packets"
  expressions: "to_int(in_broadcast_packets_raw)"
  noc_paths {
    bind: "in_broadcast_packets_raw"
    oids: "1.3.6.1.2.1.31.1.1.1.9"  # Standard MIB.
    samples: "0"
    samples: "70908519"
  }
}

transformations {
  bind: "in_multicast_packets"
  expressions: "to_int(in_multicast_packets_raw)"
  noc_paths {
    bind: "in_multicast_packets_raw"
    oids: "1.3.6.1.2.1.31.1.1.1.8"  # Standard MIB.
    samples: "0"
    samples: "402079500"
  }
}

transformations {
  bind: "in_unicast_packets"
  expressions: "to_int(in_unicast_packets_raw)"
  noc_paths {
    bind: "in_unicast_packets_raw"
    oids: "1.3.6.1.2.1.31.1.1.1.7"  # Standard MIB.
    samples: "0"
    samples: "154801769674"
  }
}

transformations {
  bind: "in_octets"
  expressions: "to_int(in_octets_raw)"
  noc_paths {
    bind: "in_octets_raw"
    oids: "1.3.6.1.2.1.31.1.1.1.6"  # Standard MIB.
    samples: "0"
    samples: "128712049996217"
  }
}

transformations {
  bind: "out_broadcast_packets"
  expressions: "to_int(out_broadcast_packets_raw)"
  noc_paths {
    bind: "out_broadcast_packets_raw"
    oids: "1.3.6.1.2.1.31.1.1.1.13"  # Standard MIB.
    samples: "0"
    samples: "72596133"
  }
}

transformations {
  bind: "out_multicast_packets"
  expressions: "to_int(out_multicast_packets_raw)"
  noc_paths {
    bind: "out_multicast_packets_raw"
    oids: "1.3.6.1.2.1.31.1.1.1.12"  # Standard MIB.
    samples: "0"
    samples: "25298767"
  }
}

transformations {
  bind: "out_unicast_packets"
  expressions: "to_int(out_unicast_packets_raw)"
  noc_paths {
    bind: "out_unicast_packets_raw"
    oids: "1.3.6.1.2.1.31.1.1.1.11"  # Standard MIB.
    samples: "0"
    samples: "153242010823"
  }
}

transformations {
  bind: "out_octets"
  expressions: "to_int(out_octets_raw)"
  noc_paths {
    bind: "out_octets_raw"
    oids: "1.3.6.1.2.1.31.1.1.1.10"  # Standard MIB.
    samples: "0"
    samples: "130289305778248"
  }
}

transformations {
  bind: "in_discards"
  expressions: "to_int(in_discards_raw)"
  noc_paths {
    bind: "in_discards_raw"
    oids: "1.3.6.1.2.1.2.2.1.13"  # Standard MIB.
    samples: "0"
  }
}

transformations {
  bind: "in_errors"
  expressions: "to_int(in_errors_raw)"
  noc_paths {
    bind: "in_errors_raw"
    oids: "1.3.6.1.2.1.2.2.1.14"  # Standard MIB.
    samples: "0"
  }
}
