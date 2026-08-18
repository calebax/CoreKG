import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Empty } from 'antd'
import AppCard from './components/AppCard'
import AddIcon from '../docs/images/add.svg?react'
import { MOCK_APPLICATIONS } from './mock'
import type { Application } from './types'
import styles from './index.module.scss'

export default function ApplicationListPage() {
  const navigate = useNavigate()
  const [apps] = useState<Application[]>(MOCK_APPLICATIONS)

  return (
    <div className={styles.page}>
      {apps.length > 0 ? (
        <div>
          <div className={styles.toolbar}>
            <Button
              className={styles.createBtn}
              onClick={() => navigate('/apps/create')}
            >
              <AddIcon className={styles.createBtnIcon} />
              创建应用
            </Button>
          </div>
          <div className={styles.grid}>
            {apps.map((app) => (
              <AppCard
                key={app.id}
                app={app}
                onClick={() => navigate(`/apps/${app.id}`)}
              />
            ))}
          </div>
        </div>
      ) : (
        <div className={styles.empty}>
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description='暂无应用'
          >
            <Button
              className={styles.createBtn}
              onClick={() => navigate('/apps/create')}
            >
              <AddIcon className={styles.createBtnIcon} />
              创建第一个应用
            </Button>
          </Empty>
        </div>
      )}
    </div>
  )
}
