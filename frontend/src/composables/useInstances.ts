import { ref, onMounted } from 'vue'
import { instancesApi } from '@/api'
import type { PGInstance } from '@/types'

export function useInstances() {
  const instances = ref<PGInstance[]>([])
  const loading   = ref(false)

  async function load(params?: { host_id?: string; status?: string }) {
    loading.value = true
    try {
      const { data } = await instancesApi.list(params)
      instances.value = data
    } finally { loading.value = false }
  }

  function labelOf(id: string) {
    const i = instances.value.find(x => x.id === id)
    return i ? `${i.instance_name}:${i.port}` : id
  }

  onMounted(() => load())
  return { instances, loading, load, labelOf }
}
