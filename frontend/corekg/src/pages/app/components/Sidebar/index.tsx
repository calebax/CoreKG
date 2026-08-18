import { SidebarWrapper } from '@/components/Layout/SidebarWrapper'
import MainSidebar from './MainSidebar'
import SidebarList from './SidebarList'
import { EExpandableStatus } from './types'

export default function Sidebar() {
  const [expandableStatus, setExpandableStatus] = useState<EExpandableStatus>(
    EExpandableStatus.EXPAND,
  )
  const handleExpandableStatusChange = () => {
    switch (expandableStatus) {
      case EExpandableStatus.FOLD:
        setExpandableStatus(EExpandableStatus.EXPAND)
        break
      case EExpandableStatus.EXPAND:
        setExpandableStatus(EExpandableStatus.FOLD)
        break
    }
  }

  return (
    <SidebarWrapper
      expandableStatus={expandableStatus}
      updateExpandableStatus={setExpandableStatus}
    >
      <MainSidebar
        expandableStatus={expandableStatus}
        onExpandableStatusChange={handleExpandableStatusChange}
      />
      <SidebarList expandableStatus={expandableStatus} />
    </SidebarWrapper>
  )
}
