import { create } from 'zustand'

interface AccessStore {
  botId: string
  spaceId: string
  setAccessIds: (botId: string, spaceId: string) => void
  clearAccessIds: () => void
}

const useAccessStore = create<AccessStore>((set) => ({
  botId: '',
  spaceId: '',
  setAccessIds: (botId, spaceId) => set({ botId, spaceId }),
  clearAccessIds: () => set({ botId: '', spaceId: '' }),
}))

export default useAccessStore
