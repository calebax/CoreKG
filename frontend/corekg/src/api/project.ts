import { send } from './request'

// 创建项目
export async function createProject(data: { [key: string]: any } = {}) {
  return send('forest.CreateProject', data)
}

// 删除项目
export async function deleteProject(data: {
  id: number
  move_to_free?: boolean
}) {
  return send('forest.DeleteProject', data)
}

// 获取项目详情
export async function getProjectInfo(data: { id: number }) {
  return send('forest.GetProjectItem', data)
}

// 获取项目列表
export async function getProjectList(data: CommonArgs) {
  const res = await send('forest.ListProjectItem', data)

  // 【临时隐藏】过滤掉智能体分组类型的项目（project_type === "agent_qa"）
  // TODO: 后期需要恢复时，删除下面的过滤逻辑即可
  if (res?.data) {
    res.data = res.data.filter((item: any) => item.project_type !== 'agent_qa')
    // 更新 total 数量
    if (res.total !== undefined) {
      res.total = res.data.length
    }
  }

  return res
}

// 重命名项目
export async function renameProject(data: { id: number; name: string }) {
  return send('forest.RenameProject', data)
}

/** 获取图表 */
export const getProjectCharts = (session_id: number) =>
  send('chat.GetChartCanvas', {
    subject_id: session_id,
    subject_type: 'session',
  })

/** 删除图表 */
export const deleteProjectCharts = (chart_ids: number[]) =>
  send('chat.BatchDeleteChart', {
    chart_ids,
  })

/** 全量更新图表 */
export const saveChartCanvas = (session_id: number, content: string) =>
  send('chat.SaveChartCanvas', {
    subject_id: session_id,
    subject_type: 'session',
    content,
  })
