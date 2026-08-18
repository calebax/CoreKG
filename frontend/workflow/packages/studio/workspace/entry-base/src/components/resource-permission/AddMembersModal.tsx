/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { useState, useEffect, useMemo } from 'react';

import {
  Modal,
  Input,
  Checkbox,
  Spin,
  Empty,
  Button,
  Toast,
} from '@coze-arch/coze-design';
import { I18n } from '@coze-arch/i18n';

import {
  CoreKGApiService,
  type EmployeeDetailInfo,
  type DepartmentInfo,
} from '@/services/corekg-api';

// 导入图标
import { ReactComponent as DepartmentIcon } from '../../../../project-entity-base/src/assets/department.svg';
import { ReactComponent as EmployeeIcon } from '../../../../project-entity-base/src/assets/employee.svg';
import { ReactComponent as DeletePersonIcon } from '../../../../project-entity-base/src/assets/deletePerson.svg';

// 员工树节点接口
interface EmployeeTreeNode {
  id: number; // 对于部门是department ID，对于员工是uin
  name: string;
  type: 'department' | 'employee';
  uin?: number; // 员工的uin
  children?: EmployeeTreeNode[];
  parentId?: number;
  departmentIds?: number[]; // 员工所属的部门ID列表
}

interface AddMembersModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (selectedIds: number[]) => void;
  initialSelectedIds?: number[];
  lockedIds?: number[];
  minSelected?: number;
}

// 树节点渲染组件
interface TreeNodeRendererProps {
  nodes: EmployeeTreeNode[];
  selectedIds: number[];
  lockedIds: number[];
  onSelectChange: (node: EmployeeTreeNode, checked: boolean) => void;
  isDepartmentFullySelected: (departmentId: number) => boolean;
  isDepartmentPartiallySelected: (departmentId: number) => boolean;
  level: number;
}

const TreeNodeRenderer: React.FC<TreeNodeRendererProps> = ({
  nodes,
  selectedIds,
  lockedIds,
  onSelectChange,
  isDepartmentFullySelected,
  isDepartmentPartiallySelected,
  level,
}) => {
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set());

  // 默认展开所有节点
  useEffect(() => {
    const allDeptIds = new Set<number>();
    const collectDeptIds = (nodeList: EmployeeTreeNode[]) => {
      nodeList.forEach(node => {
        if (node.type === 'department') {
          allDeptIds.add(node.id);
          if (node.children) {
            collectDeptIds(node.children);
          }
        }
      });
    };
    collectDeptIds(nodes);
    setExpandedIds(allDeptIds);
  }, [nodes]);

  const toggleExpand = (id: number) => {
    setExpandedIds(prev => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  };

  return (
    <div className="flex flex-col">
      {nodes.map(node => {
        const isExpanded = expandedIds.has(node.id);
        const hasChildren = node.children && node.children.length > 0;

        if (node.type === 'department') {
          const isSelected = isDepartmentFullySelected(node.id);
          const isPartiallySelected = isDepartmentPartiallySelected(node.id);
          return (
            <div key={`dept-${node.id}`}>
              <div
                className="flex items-center gap-2 px-2 py-1.5 hover:bg-[var(--coz-bg-secondary)] rounded whitespace-nowrap min-w-0"
                style={{ paddingLeft: `${level * 16 + 8}px` }}
              >
                {/* 展开/收起图标 */}
                {hasChildren && (
                  <span
                    onClick={e => {
                      e.stopPropagation();
                      toggleExpand(node.id);
                    }}
                    className="text-xs w-4 h-4 flex items-center justify-center cursor-pointer flex-shrink-0"
                  >
                    {isExpanded ? '▼' : '▶'}
                  </span>
                )}
                {!hasChildren && <span className="w-4 flex-shrink-0" />}

                {/* 复选框 */}
                <div className="flex-shrink-0">
                  <Checkbox
                    checked={isSelected}
                    indeterminate={isPartiallySelected}
                    onChange={e => {
                      e.stopPropagation();
                      onSelectChange(node, !isSelected);
                    }}
                  />
                </div>

                {/* 部门图标和名称 */}
                <div
                  className="flex items-center gap-2 flex-1 cursor-pointer min-w-0"
                  onClick={e => {
                    e.stopPropagation();
                    onSelectChange(node, !isSelected);
                  }}
                >
                  <DepartmentIcon className="w-4 h-4 flex-shrink-0" />
                  <span className="text-sm font-medium">{node.name}</span>
                </div>
              </div>

              {/* 子节点 */}
              {isExpanded && hasChildren && (
                <TreeNodeRenderer
                  nodes={node.children!}
                  selectedIds={selectedIds}
                  lockedIds={lockedIds}
                  onSelectChange={onSelectChange}
                  isDepartmentFullySelected={isDepartmentFullySelected}
                  isDepartmentPartiallySelected={isDepartmentPartiallySelected}
                  level={level + 1}
                />
              )}
            </div>
          );
        } else {
          // 员工节点
          const uin = node.uin!;
          const isSelected = selectedIds.includes(uin);
          const isLocked = lockedIds.includes(uin);

          return (
            <div
              key={`emp-${uin}`}
              className="flex items-center gap-2 px-2 py-1.5 hover:bg-[var(--coz-bg-secondary)] rounded whitespace-nowrap min-w-0"
              style={{ paddingLeft: `${level * 16 + 8}px` }}
            >
              <span className="w-4 flex-shrink-0" />
              <div className="flex-shrink-0">
                <Checkbox
                  checked={isSelected}
                  disabled={isLocked}
                  onChange={e => {
                    e.stopPropagation();
                    if (!isLocked) {
                      onSelectChange(node, !isSelected);
                    }
                  }}
                />
              </div>
              {/* 人员图标和名称 */}
              <div
                className="flex items-center gap-2 flex-1 cursor-pointer min-w-0"
                onClick={e => {
                  e.stopPropagation();
                  if (!isLocked) {
                    onSelectChange(node, !isSelected);
                  }
                }}
              >
                <EmployeeIcon className="w-4 h-4 flex-shrink-0" />
                <span className="text-sm">{node.name}</span>
              </div>
            </div>
          );
        }
      })}
    </div>
  );
};

// 将部门和员工数据转换为树结构
const convertToTree = (
  departments: DepartmentInfo[],
  employees: EmployeeDetailInfo[],
): EmployeeTreeNode[] => {
  const departmentMap = new Map<number, EmployeeTreeNode>();
  const rootNodes: EmployeeTreeNode[] = [];

  // 先创建所有部门节点
  departments.forEach(dept => {
    const node: EmployeeTreeNode = {
      id: dept.ID,
      name: dept.Name,
      type: 'department',
      children: [],
      parentId: dept.ParentID,
    };
    departmentMap.set(dept.ID, node);
  });

  // 构建部门树结构
  departments.forEach(dept => {
    const node = departmentMap.get(dept.ID);
    if (node) {
      if (dept.ParentID === 0) {
        // 根部门
        rootNodes.push(node);
      } else if (departmentMap.has(dept.ParentID)) {
        // 子部门
        departmentMap.get(dept.ParentID)?.children?.push(node);
      }
    }
  });

  // 将员工添加到对应的部门下
  employees.forEach(emp => {
    const employeeNode: EmployeeTreeNode = {
      id: emp.uin, // 员工节点使用uin作为id
      name: emp.name || emp.user_name,
      type: 'employee',
      uin: emp.uin,
      departmentIds: emp.department_ids || [],
    };

    // 如果员工有部门ID，添加到对应部门下
    if (emp.department_ids && emp.department_ids.length > 0) {
      emp.department_ids.forEach(deptId => {
        const dept = departmentMap.get(deptId);
        if (dept) {
          // 检查是否已经添加过该员工（避免重复）
          const exists = dept.children?.some(
            child => child.type === 'employee' && child.uin === emp.uin,
          );
          if (!exists) {
            dept.children?.push({ ...employeeNode });
          }
        }
      });
    } else {
      // 没有部门的员工，添加到根节点
      rootNodes.push(employeeNode);
    }
  });

  return rootNodes;
};

export const AddMembersModal: React.FC<AddMembersModalProps> = ({
  open,
  onClose,
  onConfirm,
  initialSelectedIds = [],
  lockedIds = [],
  minSelected = 0,
}) => {
  const [searchValue, setSearchValue] = useState('');
  const [selectedIds, setSelectedIds] = useState<number[]>(initialSelectedIds);
  const [loading, setLoading] = useState(false);
  const [departments, setDepartments] = useState<DepartmentInfo[]>([]);
  const [employees, setEmployees] = useState<EmployeeDetailInfo[]>([]);
  const [fetchLoading, setFetchLoading] = useState(false);

  // 获取部门树和员工列表
  const fetchDepartmentTree = async () => {
    setFetchLoading(true);
    try {
      const res = await CoreKGApiService.getDepartmentTree({
        include_employee: true,
      });
      setDepartments(res.departments || []);
      setEmployees(res.employees || []);
    } catch (error) {
      console.error('获取部门树失败:', error);
      Toast.error(I18n.t('获取部门树失败' as any, {}, '获取部门树失败'));
    } finally {
      setFetchLoading(false);
    }
  };

  useEffect(() => {
    if (open) {
      setSearchValue('');
      setSelectedIds(initialSelectedIds);
      fetchDepartmentTree();
    }
  }, [open, initialSelectedIds]);

  // 构建树结构
  const treeData = useMemo(() => {
    return convertToTree(departments, employees);
  }, [departments, employees]);

  // 过滤后的树数据（用于搜索）
  const filteredTreeData = useMemo(() => {
    if (!searchValue) return treeData;

    // 递归过滤树节点
    const filterTree = (
      nodes: EmployeeTreeNode[],
    ): EmployeeTreeNode[] | null => {
      const filtered: EmployeeTreeNode[] = [];

      nodes.forEach(node => {
        if (node.type === 'employee') {
          // 员工节点：检查名称是否匹配
          if (node.name.toLowerCase().includes(searchValue.toLowerCase())) {
            filtered.push(node);
          }
        } else {
          // 部门节点：检查部门名称或递归检查子节点
          const nameMatches = node.name
            .toLowerCase()
            .includes(searchValue.toLowerCase());
          const filteredChildren = node.children
            ? filterTree(node.children)
            : null;

          if (
            nameMatches ||
            (filteredChildren && filteredChildren.length > 0)
          ) {
            filtered.push({
              ...node,
              children: filteredChildren || node.children,
            });
          }
        }
      });

      return filtered.length > 0 ? filtered : null;
    };

    return filterTree(treeData) || [];
  }, [treeData, searchValue]);

  // 获取部门下所有员工的uin
  const getAllEmployeeUinsInDepartment = (departmentId: number): number[] => {
    const result: number[] = [];
    const collectEmployees = (nodes: EmployeeTreeNode[]) => {
      nodes.forEach(node => {
        if (node.type === 'employee' && node.uin) {
          result.push(node.uin);
        } else if (node.type === 'department' && node.children) {
          collectEmployees(node.children);
        }
      });
    };

    // 找到对应的部门节点
    const findDepartment = (
      nodes: EmployeeTreeNode[],
    ): EmployeeTreeNode | null => {
      for (const node of nodes) {
        if (node.type === 'department' && node.id === departmentId) {
          return node;
        }
        if (node.children) {
          const found = findDepartment(node.children);
          if (found) return found;
        }
      }
      return null;
    };

    const dept = findDepartment(treeData);
    if (dept && dept.children) {
      collectEmployees(dept.children);
    }
    return result;
  };

  // 检查部门是否完全选中（部门下所有员工都被选中）
  const isDepartmentFullySelected = (departmentId: number): boolean => {
    const employeeUins = getAllEmployeeUinsInDepartment(departmentId);
    return (
      employeeUins.length > 0 &&
      employeeUins.every(uin => selectedIds.includes(uin))
    );
  };

  // 检查部门是否部分选中（部门下有员工被选中，但不是全部）
  const isDepartmentPartiallySelected = (departmentId: number): boolean => {
    const employeeUins = getAllEmployeeUinsInDepartment(departmentId);
    if (employeeUins.length === 0) return false;
    const selectedCount = employeeUins.filter(uin =>
      selectedIds.includes(uin),
    ).length;
    return selectedCount > 0 && selectedCount < employeeUins.length;
  };

  // 获取已选择的成员信息
  const selectedMembers = useMemo(() => {
    return selectedIds.map(uin => {
      const emp = employees.find(e => e.uin === uin);
      return {
        id: uin,
        name: emp?.name || emp?.user_name || `用户${uin}`,
      };
    });
  }, [selectedIds, employees]);

  // 处理选择变化（支持员工和部门）
  const handleSelectChange = (node: EmployeeTreeNode, checked: boolean) => {
    if (node.type === 'employee') {
      // 员工节点：直接添加或移除该员工
      const uin = node.uin!;
      if (checked) {
        setSelectedIds(prev => {
          if (prev.includes(uin)) {
            return prev;
          }
          return [...prev, uin];
        });
      } else {
        if (lockedIds.includes(uin)) return;
        const minAllowed = Math.max(minSelected, lockedIds.length);
        const newIds = selectedIds.filter(id => id !== uin);
        if (newIds.length < minAllowed) return;
        setSelectedIds(newIds);
      }
    } else if (node.type === 'department') {
      // 部门节点：添加或移除该部门下所有员工
      const employeeUins = getAllEmployeeUinsInDepartment(node.id);
      if (checked) {
        // 选中部门：添加所有员工
        setSelectedIds(prev => {
          const newUins = employeeUins.filter(uin => !prev.includes(uin));
          return [...prev, ...newUins];
        });
      } else {
        // 取消选中部门：移除所有员工（除了被锁定的）
        const minAllowed = Math.max(minSelected, lockedIds.length);
        const newIds = selectedIds.filter(
          id => !employeeUins.includes(id) || lockedIds.includes(id),
        );
        if (newIds.length < minAllowed) return;
        setSelectedIds(newIds);
      }
    }
  };

  // 移除已选成员
  const handleRemoveSelected = (id: number) => {
    if (lockedIds.includes(id)) return;
    const minAllowed = Math.max(minSelected, lockedIds.length);
    const newIds = selectedIds.filter(i => i !== id);
    if (newIds.length < minAllowed) return;
    setSelectedIds(newIds);
  };

  // 清空已选
  const handleClearSelected = () => {
    const minAllowed = Math.max(minSelected, lockedIds.length);
    if (minAllowed === 0) {
      setSelectedIds([]);
    } else {
      setSelectedIds([...lockedIds]);
    }
  };

  // 确认添加
  const handleConfirm = () => {
    if (selectedIds.length === 0 && minSelected > 0) {
      Toast.warning(
        I18n.t('请至少选择一个成员' as any, {}, '请至少选择一个成员'),
      );
      return;
    }
    setLoading(true);
    setTimeout(() => {
      onConfirm(selectedIds);
      setLoading(false);
    }, 300);
  };

  const minAllowed = Math.max(minSelected, lockedIds.length);

  return (
    <Modal
      visible={open}
      onCancel={onClose}
      footer={null}
      width={600}
      centered
      destroyOnClose
      closable={true}
      maskClosable={false}
      closeOnEsc={false}
      title={I18n.t('选择成员' as any, {}, '选择成员')}
      bodyStyle={{ padding: '24px' }}
      getPopupContainer={() => document.body}
    >
      <div className="flex flex-col gap-4">
        {/* 内容区域 */}
        <div className="flex gap-4" style={{ height: '400px' }}>
          {/* 左侧：搜索和列表 */}
          <div className="flex flex-col gap-3 w-[260px]">
            <Input
              placeholder={I18n.t(
                '搜索部门或成员' as any,
                {},
                '搜索部门或成员',
              )}
              value={searchValue}
              onChange={setSearchValue}
              showClear
            />

            <div className="flex-1 overflow-y-auto overflow-x-auto border border-solid border-[var(--coz-stroke-primary)] rounded-lg p-2">
              {fetchLoading ? (
                <div className="flex items-center justify-center h-full">
                  <Spin />
                </div>
              ) : filteredTreeData.length === 0 ? (
                <Empty
                  description={I18n.t('暂无数据' as any, {}, '暂无数据')}
                />
              ) : (
                <TreeNodeRenderer
                  nodes={filteredTreeData}
                  selectedIds={selectedIds}
                  lockedIds={lockedIds}
                  onSelectChange={handleSelectChange}
                  isDepartmentFullySelected={isDepartmentFullySelected}
                  isDepartmentPartiallySelected={isDepartmentPartiallySelected}
                  level={0}
                />
              )}
            </div>
          </div>

          {/* 右侧：已选择成员 */}
          <div className="flex-1 p-3 rounded-lg bg-[var(--coz-bg-secondary)] flex flex-col">
            <div className="flex items-center justify-between mb-3 flex-shrink-0">
              <div className="flex items-center gap-1">
                <span className="text-sm font-medium">
                  {I18n.t('已选择' as any, {}, '已选择')}：
                </span>
                <span className="text-sm text-[var(--coz-fg-secondary)]">
                  {selectedMembers.length}
                  {I18n.t('项' as any, {}, '项')}
                </span>
              </div>
              <span
                onClick={e => {
                  e.stopPropagation();
                  if (selectedMembers.length > minAllowed) {
                    handleClearSelected();
                  }
                }}
                className={`text-sm transition-colors ${
                  selectedMembers.length > minAllowed
                    ? 'text-[var(--coz-fg-secondary)] cursor-pointer hover:text-[var(--coz-fg-primary)]'
                    : 'text-[var(--coz-fg-dim)] cursor-not-allowed'
                }`}
              >
                {I18n.t('清空已选' as any, {}, '清空已选')}
              </span>
            </div>

            <div className="flex-1 overflow-y-auto">
              {selectedMembers.length > 0 ? (
                <div className="flex flex-col gap-1">
                  {selectedMembers.map(member => {
                    const isLocked = lockedIds.includes(member.id);
                    const disableDelete =
                      isLocked || selectedMembers.length <= minAllowed;
                    return (
                      <div
                        key={member.id}
                        className="flex items-center justify-between px-2 py-1.5 hover:bg-[var(--coz-bg-tertiary)] rounded"
                      >
                        <div className="flex items-center gap-2 flex-1">
                          {/* 人员图标 */}
                          <EmployeeIcon className="w-4 h-4 flex-shrink-0" />
                          <span className="text-sm">{member.name}</span>
                        </div>
                        <DeletePersonIcon
                          onClick={e => {
                            e.stopPropagation();
                            if (!disableDelete) {
                              handleRemoveSelected(member.id);
                            }
                          }}
                          className={`w-4 h-4 flex-shrink-0 transition-opacity ${
                            disableDelete
                              ? 'opacity-30 cursor-not-allowed'
                              : 'cursor-pointer hover:opacity-80'
                          }`}
                        />
                      </div>
                    );
                  })}
                </div>
              ) : (
                <div className="text-center text-[var(--coz-fg-secondary)] py-8">
                  {I18n.t('暂无已选成员' as any, {}, '暂无已选成员')}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* 底部按钮 */}
        <div className="flex justify-end gap-2">
          <Button
            onClick={e => {
              e.stopPropagation();
              onClose();
            }}
            style={{
              padding: '9px 24.5px',
              border: 'none',
              borderRadius: '6px',
              backgroundColor: '#F5F5F5',
              color: '#0C1F17',
            }}
          >
            {I18n.t('Cancel')}
          </Button>
          <Button
            type="primary"
            loading={loading}
            onClick={e => {
              e.stopPropagation();
              handleConfirm();
            }}
            style={{
              padding: '9px 24.5px',
              border: 'none',
              borderRadius: '6px',
              backgroundColor: loading ? 'rgba(0, 0, 0, 0.06)' : '#0C99FF',
              color: loading ? 'rgba(0, 0, 0, 0.25)' : '#ffffff',
            }}
          >
            {I18n.t('确定' as any, {}, '确定')}
          </Button>
        </div>
      </div>
    </Modal>
  );
};

export default AddMembersModal;
