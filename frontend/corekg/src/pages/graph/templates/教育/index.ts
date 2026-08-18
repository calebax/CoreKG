import { GraphTag, GraphTagRelationship, GraphTemplate } from 'Graph'

// 实体类型（GraphTag）
export const graphTags: GraphTag[] = [
  {
    tag_name: '学生',
    properties: [
      { name: '学号', comment: 'student_id', type: 'string' },
      { name: '姓名', comment: 'student_name', type: 'string' },
      { name: '年级', comment: 'grade', type: 'string' },
      { name: '是否在读', comment: 'is_active', type: 'string' },
    ],
  },
  {
    tag_name: '教师',
    properties: [
      { name: '工号', comment: 'teacher_id', type: 'string' },
      { name: '姓名', comment: 'teacher_name', type: 'string' },
      { name: '职称', comment: 'title', type: 'string' },
      { name: '学院', comment: 'department', type: 'string' },
    ],
  },
  {
    tag_name: '课程',
    properties: [
      { name: '课程编号', comment: 'course_id', type: 'string' },
      { name: '课程名称', comment: 'course_name', type: 'string' },
      { name: '学分', comment: 'credits', type: 'string' },
      { name: '学期', comment: 'term', type: 'string' },
    ],
  },
  {
    tag_name: '章节',
    properties: [
      { name: '章节编号', comment: 'chapter_id', type: 'string' },
      { name: '标题', comment: 'title', type: 'string' },
      { name: '时长(分钟)', comment: 'duration_min', type: 'string' },
      { name: '是否必修', comment: 'is_required', type: 'string' },
    ],
  },
  {
    tag_name: '作业',
    properties: [
      { name: '作业编号', comment: 'assignment_id', type: 'string' },
      { name: '标题', comment: 'title', type: 'string' },
      { name: '发布日期', comment: 'publish_date', type: 'string' },
      { name: '截止日期', comment: 'due_date', type: 'string' },
    ],
  },
  {
    tag_name: '考试',
    properties: [
      { name: '考试编号', comment: 'exam_id', type: 'string' },
      { name: '考试名称', comment: 'exam_name', type: 'string' },
      { name: '考试时间', comment: 'exam_time', type: 'string' },
      { name: '总分', comment: 'total_score', type: 'string' },
    ],
  },
  {
    tag_name: '题目',
    properties: [
      { name: '题目编号', comment: 'question_id', type: 'string' },
      { name: '题型', comment: 'question_type', type: 'string' },
      { name: '分值', comment: 'score', type: 'string' },
      { name: '难度', comment: 'difficulty', type: 'string' },
    ],
  },
  {
    tag_name: '证书',
    properties: [
      { name: '证书编号', comment: 'certificate_id', type: 'string' },
      { name: '证书名称', comment: 'certificate_name', type: 'string' },
      { name: '颁发日期', comment: 'issue_date', type: 'string' },
    ],
  },
]

// 关系（GraphTagRelationship）
export const graphTagRelationships: GraphTagRelationship[] = [
  { name: '任教', source: '教师', target: '课程' },
  { name: '选修', source: '学生', target: '课程' },
  { name: '包含章节', source: '课程', target: '章节' },
  { name: '布置', source: '课程', target: '作业' },
  { name: '参加考试', source: '学生', target: '考试' },
  { name: '包含题目', source: '考试', target: '题目' },
  { name: '获得', source: '学生', target: '证书' },
]
const Template: GraphTemplate = {
  name: '教育在线学习模板',
  description:
    '围绕“课程、班级、学员、学习进度/测评”的关系，清晰展示课程结构、学习路径与成绩评价。',
  // avatar,
  tags: graphTags,
  edges: graphTagRelationships,
}
export default Template
