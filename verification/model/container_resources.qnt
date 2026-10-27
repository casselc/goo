// Bounded observed macrosteps, not arbitrary UI schedules.
// active/pendingMetadata are internal guards; Go validates actual withdrawals.
module container_resources {
 const MAX_CACHE: int
 var reads: int
 var plans: int
 var memoryComposes: int
 var cpuComposes: int
 var cached: int
 var memoryLoaded: bool
 var cpuLoaded: bool
 var pendingMemory: bool
 var pendingCPU: bool
 var pendingMetadata: bool
 var active: bool
 action init = all {reads'=0,plans'=0,memoryComposes'=0,cpuComposes'=0,cached'=0,memoryLoaded'=false,cpuLoaded'=false,pendingMemory'=false,pendingCPU'=false,pendingMetadata'=false,active'=true}
 action update(nreads: int,nplans: int,nmemoryComposes: int,ncpuComposes: int,ncached: int,nmemoryLoaded: bool,ncpuLoaded: bool,npendingMemory: bool,npendingCPU: bool,npendingMetadata: bool,nactive: bool): bool = all {reads'=nreads,plans'=nplans,memoryComposes'=nmemoryComposes,cpuComposes'=ncpuComposes,cached'=ncached,memoryLoaded'=nmemoryLoaded,cpuLoaded'=ncpuLoaded,pendingMemory'=npendingMemory,pendingCPU'=npendingCPU,pendingMetadata'=npendingMetadata,active'=nactive}
 action inventory = all { reads==0, update(1,plans,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,pendingMetadata,true) }
 action discover = all { active and cached<MAX_CACHE, update(reads+1,plans+1,memoryComposes,cpuComposes,cached+1,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,false,active) }
 action memory_load = all { active and cached>0, update(reads+3,plans,memoryComposes+1,cpuComposes,cached,true,cpuLoaded,false,pendingCPU,pendingMetadata,active) }
 action cpu_load = all { active and cached>0, update(reads+3,plans,memoryComposes,cpuComposes+1,cached,memoryLoaded,true,pendingMemory,false,pendingMetadata,active) }
 action native_stage = all { active, update(reads,plans+1,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,pendingMetadata,false) }
 action back = all { not(active), update(reads,plans,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,pendingMetadata,true) }
 action native_run = all { active, update(reads+1,plans+1,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,pendingMetadata,false) }
 action capture = all { not(active), update(reads+1,plans,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,pendingMetadata,active) }
 action aligned = all { true, update(reads+7,plans+1,memoryComposes+1,cpuComposes+1,1,true,true,pendingMemory,pendingCPU,pendingMetadata,true) }
 action freeze = all { true, update(reads,plans,memoryComposes,cpuComposes,0,false,false,pendingMemory,pendingCPU,pendingMetadata,false) }
 action start_cpu = all { active and cached>0, update(reads+1,plans,memoryComposes,cpuComposes,cached,memoryLoaded,false,pendingMemory,true,pendingMetadata,active) }
 action start_memory = all { active and cached>0, update(reads+1,plans,memoryComposes,cpuComposes,cached,false,cpuLoaded,true,pendingCPU,pendingMetadata,active) }
 action source_edit = all { active, update(reads,plans,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,false,active) }
 action metadata_changes_memory = all { active and cached<MAX_CACHE, update(reads+1,plans+1,memoryComposes,cpuComposes,cached+1,false,cpuLoaded,false,pendingCPU,pendingMetadata,active) }
 action edit_memory = all { active, update(reads,plans,memoryComposes,cpuComposes,cached,false,cpuLoaded,false,pendingCPU,pendingMetadata,active) }
 action edit_cpu = all { active, update(reads,plans,memoryComposes,cpuComposes,cached,memoryLoaded,false,pendingMemory,false,pendingMetadata,active) }
 action close = all { active, update(reads,plans,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,false,false,false,false) }
 action finish_cpu = all { active and pendingCPU, update(reads,plans,memoryComposes,cpuComposes+1,cached,memoryLoaded,true,pendingMemory,false,pendingMetadata,active) }
 action late_cpu = all { not(pendingCPU), update(reads,plans,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,pendingMetadata,active) }
 action late_memory = all { not(pendingMemory), update(reads,plans,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,pendingMetadata,active) }
 action start_metadata = all { active, update(reads+1,plans,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,true,active) }
 action late_metadata = all { not(pendingMetadata), update(reads,plans,memoryComposes,cpuComposes,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,pendingMetadata,active) }
 action begin_compose_cpu = all { active and pendingCPU, update(reads,plans,memoryComposes,cpuComposes+1,cached,memoryLoaded,cpuLoaded,pendingMemory,pendingCPU,pendingMetadata,active) }
 val safety = reads>=0 and plans>=0 and memoryComposes>=0 and cpuComposes>=0 and cached>=0 and cached<=MAX_CACHE and not(memoryLoaded and pendingMemory) and not(cpuLoaded and pendingCPU)
}
