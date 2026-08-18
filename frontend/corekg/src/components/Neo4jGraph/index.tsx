/*
 * Copyright (c) "Neo4j"
 * Neo4j Sweden AB [http://neo4j.com]
 *
 * This file is part of Neo4j.
 *
 * Neo4j is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
import React from 'react'
import { ResizeObserver } from '@juggle/resize-observer'
import {
  GraphEventHandlerModel,
  GraphInteractionCallBack,
} from './GraphEventHandlerModel'
import {
  BasicNode,
  BasicRelationship,
  ZoomInIcon,
  ZoomOutIcon,
  ZoomToFitIcon,
} from './common'
import { GraphModel } from './models/Graph'
import { GraphStyleModel } from './models/GraphStyle'
import { NodeModel } from './models/Node'
import {
  StyledSvgWrapper,
  StyledZoomButton,
  StyledZoomHolder,
  ThemeProvider,
} from './styled'
import {
  GetNodeNeighboursFn,
  VizItem,
  ZoomLimitsReached,
  ZoomType,
} from './types'
import {
  GraphStats,
  createGraph,
  getGraphStats,
  mapNodes,
  mapRelationships,
} from './utils/mapper'
import { Visualization } from './visualization/Visualization'

export type OperatorClickType = 'editNode' | 'createEdge' | 'deleteNode'

/** 节点操作 Action */
export type NodeAction =
  | { type: 'add'; nodeData: BasicNode | BasicNode[] }
  | { type: 'edit'; nodeId: string; nodeData: BasicNode }
  | { type: 'delete'; nodeId: string }

/** 边操作 Action */
export type EdgeAction =
  | { type: 'add'; edgeData: BasicRelationship | BasicRelationship[] }
  | { type: 'edit'; edgeId: string; edgeData: BasicRelationship }
  | { type: 'delete'; edgeId: string }

export type GraphProps = {
  isFullscreen?: boolean
  /** 变更后会刷新图 */
  graphStyle?: GraphStyleModel
  /** 初始的节点与关系 变化后不会更新图 */
  relationships: BasicRelationship[]
  nodes: BasicNode[]
  /** 获取邻居节点及之间的关系 */
  getNodeNeighbours: GetNodeNeighboursFn
  /** 图交互的事件 包括点击双击等 */
  onGraphInteraction?: GraphInteractionCallBack
  /** 图的节点、边相关事件 */
  onItemMouseOver?: (item: VizItem) => void
  onItemSelect?: (item: VizItem) => void
  /** 图数据变化的回调函数 初次构建、增减边和节点均会触发 */
  onGraphModelChange?: (stats: GraphStats) => void
  /** 点击节点操作器的回调函数，支持三种操作类型：editNode、createEdge、deleteNode */
  onClickOperators?: (
    type: OperatorClickType,
    node: NodeModel,
    endNode?: NodeModel,
  ) => void
  /** 是否显示节点操作器 */
  showNodeOperators?: boolean
  /** 是否只读模式（禁用节点操作） */
  isReadOnly?: boolean
  /** 缩放按钮的右边距 */
  offset?: number
}

type GraphState = {
  zoomInLimitReached: boolean
  zoomOutLimitReached: boolean
}

export default class Neo4jGraph extends React.Component<
  GraphProps,
  GraphState
> {
  svgElement: React.RefObject<SVGSVGElement>
  wrapperElement: React.RefObject<HTMLDivElement>
  wrapperResizeObserver: ResizeObserver
  visualization: Visualization | null = null
  graph: GraphModel | null = null

  constructor(props: GraphProps) {
    super(props)
    this.state = {
      zoomInLimitReached: false,
      zoomOutLimitReached: false,
    }
    this.svgElement = React.createRef()
    this.wrapperElement = React.createRef()

    this.wrapperResizeObserver = new ResizeObserver(() => {
      this.visualization?.resize(this.props.isFullscreen)
    })
  }

  componentDidMount(): void {
    const {
      getNodeNeighbours,
      isFullscreen,
      graphStyle = new GraphStyleModel(true),
      nodes,
      onGraphInteraction,
      onGraphModelChange,
      onItemMouseOver,
      onItemSelect,
      onClickOperators,
      relationships,
      showNodeOperators,
      isReadOnly,
    } = this.props

    if (!this.svgElement.current) return

    const measureSize = () => ({
      width: this.svgElement.current?.parentElement?.clientWidth ?? 200,
      height: this.svgElement.current?.parentElement?.clientHeight ?? 200,
    })

    const graph = createGraph(nodes, relationships)
    this.graph = graph
    this.visualization = new Visualization(
      this.svgElement.current,
      measureSize,
      this.handleZoomEvent,
      graph,
      graphStyle,
      isFullscreen,
      showNodeOperators,
      isReadOnly,
    )
    const graphEventHandler = new GraphEventHandlerModel(
      graph,
      this.visualization,
      getNodeNeighbours,
      onItemMouseOver,
      onItemSelect,
      onGraphModelChange,
      onGraphInteraction,
      onClickOperators,
    )
    graphEventHandler.bindEventHandlers()
    onGraphModelChange?.(getGraphStats(graph))
    this.visualization.resize(isFullscreen)

    this.visualization?.init()
    this.visualization?.precomputeAndStart()

    this.wrapperResizeObserver.observe(this.svgElement.current)
  }

  componentDidUpdate(prevProps: GraphProps): void {
    if (this.props.isFullscreen !== prevProps.isFullscreen) {
      this.visualization?.resize(this.props.isFullscreen)
    }
    if (this.props.graphStyle !== prevProps.graphStyle && this.visualization) {
      this.visualization.style =
        this.props.graphStyle ?? new GraphStyleModel(true)
      this.visualization.update({
        updateNodes: true,
        updateRelationships: true,
        restartSimulation: false,
      })
    }
    if (
      (this.props.showNodeOperators !== prevProps.showNodeOperators ||
        this.props.isReadOnly !== prevProps.isReadOnly) &&
      this.visualization
    ) {
      this.visualization.showNodeOperators = this.props.showNodeOperators
      this.visualization.isReadOnly = this.props.isReadOnly
      this.visualization.update({
        updateNodes: true,
        updateRelationships: false,
        restartSimulation: false,
      })
    }
  }

  componentWillUnmount(): void {
    this.wrapperResizeObserver.disconnect()
  }

  handleZoomEvent = (limitsReached: ZoomLimitsReached): void => {
    if (
      limitsReached.zoomInLimitReached !== this.state.zoomInLimitReached ||
      limitsReached.zoomOutLimitReached !== this.state.zoomOutLimitReached
    ) {
      this.setState({
        zoomInLimitReached: limitsReached.zoomInLimitReached,
        zoomOutLimitReached: limitsReached.zoomOutLimitReached,
      })
    }
  }

  zoomInClicked = (): void => {
    this.visualization?.zoomByType(ZoomType.IN)
  }

  zoomOutClicked = (): void => {
    this.visualization?.zoomByType(ZoomType.OUT)
  }

  zoomToFitClicked = (): void => {
    this.visualization?.zoomByType(ZoomType.FIT)
  }

  /** 节点操作：增删改 */
  dispatchNodeAction = (action: NodeAction): void => {
    if (!this.graph || !this.visualization) return

    switch (action.type) {
      case 'add': {
        const nodes = Array.isArray(action.nodeData)
          ? action.nodeData
          : [action.nodeData]
        const nodeModels = mapNodes(nodes)
        this.graph.addNodes(nodeModels)
        this.visualization.update({
          updateNodes: true,
          updateRelationships: false,
          restartSimulation: true,
        })
        this.props.onGraphModelChange?.(getGraphStats(this.graph))
        break
      }

      case 'edit': {
        const node = this.graph.findNode(action.nodeId)
        if (!node) return

        // 先删除旧节点，再添加新节点
        this.graph.removeConnectedRelationships(node)
        this.graph.removeNode(node)

        const nodeModels = mapNodes([action.nodeData])
        this.graph.addNodes(nodeModels)
        this.visualization.update({
          updateNodes: true,
          updateRelationships: true,
          restartSimulation: true,
        })
        this.props.onGraphModelChange?.(getGraphStats(this.graph))
        break
      }

      case 'delete': {
        const node = this.graph.findNode(action.nodeId)
        if (!node) return

        // 删除节点及其相关关系
        this.graph.removeConnectedRelationships(node)
        this.graph.removeNode(node)

        this.visualization.update({
          updateNodes: true,
          updateRelationships: true,
          restartSimulation: true,
        })
        this.props.onGraphModelChange?.(getGraphStats(this.graph))
        break
      }
    }
  }

  /** 边操作：增删改 */
  dispatchEdgeAction = (action: EdgeAction): void => {
    if (!this.graph || !this.visualization) return

    switch (action.type) {
      case 'add': {
        const relationships = Array.isArray(action.edgeData)
          ? action.edgeData
          : [action.edgeData]
        const relationshipModels = mapRelationships(relationships, this.graph)
        this.graph.addRelationships(relationshipModels)
        this.visualization.update({
          updateNodes: false,
          updateRelationships: true,
          restartSimulation: true,
        })
        this.props.onGraphModelChange?.(getGraphStats(this.graph))
        break
      }

      case 'edit': {
        const relationship = this.graph.findRelationship(action.edgeId)
        if (!relationship) return

        // 更新源节点和目标节点
        this.graph.updateNode(relationship.source)
        this.graph.updateNode(relationship.target)

        // 删除旧关系
        const relationships = this.graph.relationships()
        const index = relationships.indexOf(relationship)
        if (index !== -1) {
          relationships.splice(index, 1)
          delete this.graph.relationshipMap[action.edgeId]
        }

        // 添加新关系
        const relationshipModels = mapRelationships(
          [action.edgeData],
          this.graph,
        )
        this.graph.addRelationships(relationshipModels)
        this.visualization.update({
          updateNodes: true,
          updateRelationships: true,
          restartSimulation: true,
        })
        this.props.onGraphModelChange?.(getGraphStats(this.graph))
        break
      }

      case 'delete': {
        const relationship = this.graph.findRelationship(action.edgeId)
        if (!relationship) return

        // 更新源节点和目标节点
        this.graph.updateNode(relationship.source)
        this.graph.updateNode(relationship.target)

        // 删除关系：从 relationships 数组和 relationshipMap 中移除
        const relationships = this.graph.relationships()
        const index = relationships.indexOf(relationship)
        if (index !== -1) {
          relationships.splice(index, 1)
          delete this.graph.relationshipMap[action.edgeId]
        }

        this.visualization.update({
          updateNodes: true,
          updateRelationships: true,
          restartSimulation: true,
        })
        this.props.onGraphModelChange?.(getGraphStats(this.graph))
        break
      }
    }
  }

  /** 获取图谱统计信息 */
  getGraphStats = (): GraphStats | undefined => {
    if (!this.graph) return undefined
    return getGraphStats(this.graph)
  }

  /** 重置图谱数据并刷新渲染（用于外部强制刷新） */
  resetGraphData = (
    nodes: BasicNode[],
    relationships: BasicRelationship[],
    options?: { zoomToFit?: boolean },
  ): void => {
    if (!this.graph || !this.visualization) return

    this.graph.resetGraph()
    this.graph.addNodes(mapNodes(nodes))
    this.graph.addRelationships(mapRelationships(relationships, this.graph))

    this.visualization.update({
      updateNodes: true,
      updateRelationships: true,
      restartSimulation: true,
    })

    this.props.onGraphModelChange?.(getGraphStats(this.graph))

    if (options?.zoomToFit) {
      this.visualization.zoomByType(ZoomType.FIT)
    }
  }

  /** 缩放至适配视口 */
  zoomToFit = (): void => {
    this.visualization?.zoomByType(ZoomType.FIT)
  }

  render(): JSX.Element {
    const { offset = 8, isFullscreen } = this.props
    const { zoomInLimitReached, zoomOutLimitReached } = this.state
    return (
      <ThemeProvider>
        <StyledSvgWrapper ref={this.wrapperElement}>
          <svg ref={this.svgElement} />
          <StyledZoomHolder offset={offset} isFullscreen={isFullscreen}>
            <StyledZoomButton
              aria-label={'zoom-in'}
              className={'zoom-in'}
              disabled={zoomInLimitReached}
              onClick={this.zoomInClicked}
            >
              <ZoomInIcon large={isFullscreen} />
            </StyledZoomButton>
            <StyledZoomButton
              aria-label={'zoom-out'}
              className={'zoom-out'}
              disabled={zoomOutLimitReached}
              onClick={this.zoomOutClicked}
            >
              <ZoomOutIcon large={isFullscreen} />
            </StyledZoomButton>
            <StyledZoomButton
              aria-label={'zoom-to-fit'}
              onClick={this.zoomToFitClicked}
            >
              <ZoomToFitIcon large={isFullscreen} />
            </StyledZoomButton>
          </StyledZoomHolder>
        </StyledSvgWrapper>
      </ThemeProvider>
    )
  }
}
