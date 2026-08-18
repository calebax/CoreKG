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
import { BaseType } from 'd3-selection'
import { NodeCaptionLine, NodeModel } from '../../models/Node'
import { RelationshipModel } from '../../models/Relationship'
import Renderer from '../Renderer'

const noop = () => undefined

const nodeRingStrokeSize = 8

const nodeOutline = new Renderer<NodeModel>({
  name: 'nodeOutline',
  onGraphChange(selection, viz) {
    return selection
      .selectAll('circle.b-outline')
      .data((node) => [node])
      .join('circle')
      .classed('b-outline', true)
      .attr('cx', 0)
      .attr('cy', 0)
      .attr('r', (node: NodeModel) => {
        return node.radius
      })
      .attr('fill', (node: NodeModel) => {
        return viz.style.forNode(node).get('color')
      })
      .attr('stroke', (node: NodeModel) => {
        return viz.style.forNode(node).get('border-color')
      })
      .attr('stroke-width', (node: NodeModel) => {
        return viz.style.forNode(node).get('border-width')
      })
  },
  onTick: noop,
})

const nodeCaption = new Renderer<NodeModel>({
  name: 'nodeCaption',
  onGraphChange(selection, viz) {
    return (
      selection
        .selectAll('text.caption')
        .data((node: NodeModel) => node.caption)
        .join('text')
        // Classed element ensures duplicated data will be removed before adding
        .classed('caption', true)
        .attr('text-anchor', 'middle')
        .attr('pointer-events', 'none')
        .attr('x', 0)
        .attr('y', (line: NodeCaptionLine) => line.baseline)
        .attr('font-size', (line: NodeCaptionLine) =>
          viz.style.forNode(line.node).get('font-size'),
        )
        .attr('fill', (line: NodeCaptionLine) =>
          viz.style.forNode(line.node).get('text-color-internal'),
        )
        .text((line: NodeCaptionLine) => line.text)
    )
  },

  onTick: noop,
})

const nodeRing = new Renderer<NodeModel>({
  name: 'nodeRing',
  onGraphChange(selection) {
    const circles = selection
      .selectAll('circle.ring')
      .data((node: NodeModel) => [node])

    const enter = circles.enter()

    enter
      .insert('circle', '.b-outline')
      .classed('ring', true)
      .attr('cx', 0)
      .attr('cy', 0)
      .attr('stroke-width', `${nodeRingStrokeSize}px`)
      .attr('r', (node: NodeModel) => node.radius + 4)

    return circles.exit().remove()
  },

  onTick: noop,
})

const nodeOperators = new Renderer<NodeModel>({
  name: 'nodeOperators',
  onGraphChange(selection, viz) {
    const operatorConfig = {
      // 容器配置
      container: {
        offsetX: 55, // 容器向右偏移距离
      },
      // 按钮尺寸
      button: {
        width: 38,
        height: 18,
        borderRadius: 2,
        spacing: 21,
      },
      // 按钮样式配置
      style: {
        backgroundColor: '#fff', // 按钮背景色
        strokeWidth: 1, // 边框宽度
      },
      // 文字配置
      text: {
        fontSize: 8, // 字体大小
      },
      // 按钮内容配置
      buttons: [
        {
          text: '编辑实体',
          eventType: 'clickOperator:editNode',
          strokeColor: '#4A90E2', // 边框和文字颜色
        },
        {
          text: '新建关系',
          eventType: 'clickOperator:startCreateEdge',
          strokeColor: '#4A90E2',
        },
        {
          text: '快速删除',
          eventType: 'clickOperator:deleteNode',
          strokeColor: '#9B9B9B',
        },
      ],
    }

    // 计算按钮位置（基于配置）
    const buttonRectX = -operatorConfig.button.width / 2 // 矩形x位置（水平居中）
    const buttonRectY = -operatorConfig.button.height / 2 // 矩形y位置（垂直居中）
    const buttonTextY =
      operatorConfig.button.height / 2 - operatorConfig.text.fontSize / 2 - 1 // 文本y位置（垂直居中，考虑字体基线）

    // 计算每个按钮的y位置
    const buttonPositions = operatorConfig.buttons.map((_, index) => {
      const centerIndex = Math.floor(operatorConfig.buttons.length / 2)
      const offset = (index - centerIndex) * operatorConfig.button.spacing
      return offset
    })

    // ============= 渲染逻辑 =============

    if (viz.edgeStartNode || !viz.showNodeOperators) {
      return selection.selectAll('g.node-operators').remove()
    }

    const operatorsGroup = selection
      .selectAll('g.node-operators')
      .data((node: NodeModel) => [node])
      .on('click', (e) => {
        e.stopPropagation()
      })
      .on('dblclick', (e) => {
        e.stopPropagation()
      })

    const enter = operatorsGroup.enter()

    const group = enter
      .insert('g', '.b-outline')
      .classed('node-operators', true)
      .attr('transform', `translate(${operatorConfig.container.offsetX}, 0)`)

    // 容器尺寸 宽度额外增加offsetX以扩大hover区域
    const containerWidth =
      operatorConfig.button.width + operatorConfig.container.offsetX
    const containerHeight =
      (operatorConfig.buttons.length - 1) * operatorConfig.button.spacing +
      operatorConfig.button.height
    const containerTop = -containerHeight / 2 + operatorConfig.button.height / 2

    // 添加透明背景矩形，扩大hover区域
    group
      .append('rect')
      .attr('x', -containerWidth / 2)
      .attr('y', containerTop - operatorConfig.button.height / 2)
      .attr('width', containerWidth)
      .attr('height', containerHeight)
      .attr('fill', 'transparent')
      .attr('pointer-events', 'all')
      .style('cursor', 'default')

    // 只读模式下的禁用样式
    const isReadOnly = viz.isReadOnly ?? false
    const disabledColor = '#C0C4CC' // 禁用状态的颜色

    operatorConfig.buttons.forEach((buttonConfig, index) => {
      const button = group
        .append('g')
        .classed('operator-button', true)
        .classed('btn', true)
        .attr('transform', `translate(0, ${buttonPositions[index]})`)
        .style('cursor', isReadOnly ? 'not-allowed' : 'pointer')
        .on('click', (e, d) => {
          e.stopPropagation()
          if (!isReadOnly) {
            viz.trigger(buttonConfig.eventType, d)
          }
        })

      // 根据是否只读选择颜色
      const buttonStrokeColor = isReadOnly
        ? disabledColor
        : buttonConfig.strokeColor
      const buttonTextColor = isReadOnly ? disabledColor : buttonConfig.strokeColor

      button
        .append('rect')
        .attr('x', buttonRectX)
        .attr('y', buttonRectY)
        .attr('width', operatorConfig.button.width)
        .attr('height', operatorConfig.button.height)
        .attr('rx', operatorConfig.button.borderRadius)
        .attr('fill', operatorConfig.style.backgroundColor)
        .attr('stroke', buttonStrokeColor)
        .attr('stroke-width', operatorConfig.style.strokeWidth)
        .attr('opacity', isReadOnly ? 0.5 : 1) // 只读时降低透明度

      // 按钮文本
      button
        .append('text')
        .text(buttonConfig.text)
        .attr('x', 0)
        .attr('y', buttonTextY)
        .attr('text-anchor', 'middle')
        .attr('font-size', operatorConfig.text.fontSize)
        .attr('fill', buttonTextColor)
        .attr('pointer-events', 'none')
    })

    return operatorsGroup.exit().remove()
  },

  onTick: noop,
})

const arrowPath = new Renderer<RelationshipModel>({
  name: 'arrowPath',

  onGraphChange(selection, viz) {
    return selection
      .selectAll('path.b-outline')
      .data((rel: any) => [rel])
      .join('path')
      .classed('b-outline', true)
      .attr('fill', (rel: any) => viz.style.forRelationship(rel).get('color'))
      .attr('stroke', 'none')
  },

  onTick(selection) {
    return selection
      .selectAll<BaseType, RelationshipModel>('path')
      .attr('d', (d) => d.arrow!.outline(d.shortCaptionLength ?? 0))
  },
})

const relationshipType = new Renderer<RelationshipModel>({
  name: 'relationshipType',
  onGraphChange(selection, viz) {
    return selection
      .selectAll('text')
      .data((rel) => [rel])
      .join('text')
      .attr('text-anchor', 'middle')
      .attr('pointer-events', 'none')
      .attr('font-size', (rel) =>
        viz.style.forRelationship(rel).get('font-size'),
      )
      .attr('fill', (rel) =>
        viz.style.forRelationship(rel).get(`text-color-${rel.captionLayout}`),
      )
  },

  onTick(selection, viz) {
    return selection
      .selectAll<BaseType, RelationshipModel>('text')
      .attr('x', (rel) => rel?.arrow?.midShaftPoint?.x ?? 0)
      .attr(
        'y',
        (rel) =>
          (rel?.arrow?.midShaftPoint?.y ?? 0) +
          parseFloat(viz.style.forRelationship(rel).get('font-size')) / 2 -
          1,
      )
      .attr('transform', (rel) => {
        if (rel.naturalAngle < 90 || rel.naturalAngle > 270) {
          return `rotate(180 ${rel?.arrow?.midShaftPoint?.x ?? 0} ${
            rel?.arrow?.midShaftPoint?.y ?? 0
          })`
        } else {
          return null
        }
      })
      .text((rel) => rel.shortCaption ?? '')
  },
})

const relationshipOverlay = new Renderer<RelationshipModel>({
  name: 'relationshipOverlay',
  onGraphChange(selection) {
    return selection
      .selectAll('path.overlay')
      .data((rel) => [rel])
      .join('path')
      .classed('overlay', true)
  },

  onTick(selection) {
    const band = 16

    return selection
      .selectAll<BaseType, RelationshipModel>('path.overlay')
      .attr('d', (d) => d.arrow!.overlay(band))
  },
})

const node = [nodeOutline, nodeCaption, nodeRing, nodeOperators]

const relationship = [arrowPath, relationshipType, relationshipOverlay]

export { node, relationship }
