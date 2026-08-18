import { FC, useState } from 'react'
import ApiAccessSection from './components/ApiAccessSection'
import EmbedCodeModal from './components/EmbedCodeModal'
import EmbedSection from './components/EmbedSection'
import PublicAccessSection from './components/PublicAccessSection'

const AgentAccess: FC = () => {
  const [embedModalVisible, setEmbedModalVisible] = useState(false)

  return (
    <div className='p-8'>
      <PublicAccessSection />

      <div className='mt-8'>
        <EmbedSection onViewCode={() => setEmbedModalVisible(true)} />
      </div>

      <div className='mt-8 pt-6 border-t border-gray-200'>
        <ApiAccessSection />
      </div>

      <EmbedCodeModal
        visible={embedModalVisible}
        onClose={() => setEmbedModalVisible(false)}
      />
    </div>
  )
}

export default AgentAccess
