import { Empty, Skeleton } from 'antd'

export default function AnalyticsTab() {
  return (
    <div style={{ padding: '40px 0' }}>
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description={
          <>
            <Skeleton active paragraph={{ rows: 3 }} />
            <div style={{ marginTop: 8, color: '#919497' }}>
              运营分析功能开发中
            </div>
          </>
        }
      />
    </div>
  )
}
