import React, { useState, useEffect } from 'react'
import { App, Button, Form, Input, message } from 'antd'
import { useTranslation } from 'react-i18next'
import { cn } from '@/utils'
import {
  chooseUin,
  createCompany,
  loginByPassword,
  loginByPasswordPrivate,
} from '@/api/account'
import { Agreement } from '@/components/agreement'
import useLocalStore from '@/stores/local'
import { encryptPassword } from '@/utils/crypto'
import { useDeployConfig } from '@/utils/useDeployConfig'

interface CreateCompanyProps {
  info: {
    refreshToken: string
    userInfo: any
    uinList: any[]
    username?: string
    password?: string
  }
  onLogin: () => void
}

const CreateCompany: React.FC<CreateCompanyProps> = ({ info, onLogin }) => {
  const { version } = useDeployConfig()
  const { t } = useTranslation('pages')
  const { t: tM } = useTranslation('messages')
  const { message: messageApi } = App.useApp()
  const localStore = useLocalStore()
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  // 监听表单字段值，判断按钮是否可点击
  const companyName = Form.useWatch('company_name', form)
  const userDisplayName = Form.useWatch('user_display_name', form)
  const isButtonDisabled =
    !companyName?.trim() || !userDisplayName?.trim() || loading

  // 微信登录时，将微信名称回显到用户昵称输入框
  useEffect(() => {
    // 只有在微信登录（没有用户名密码）且微信名称存在时才回显
    if (!info.username && info.userInfo?.name) {
      const currentUserDisplayName = form.getFieldValue('user_display_name')
      // 如果输入框为空，则设置微信名称
      if (!currentUserDisplayName?.trim()) {
        form.setFieldValue('user_display_name', info.userInfo.name)
      }
    }
  }, [info.userInfo?.name, info.username, form])

  const handleCreateAndLogin = async (values: {
    company_name: string
    user_display_name: string
  }) => {
    // 如果按钮不可用，阻止提交
    if (isButtonDisabled) {
      return
    }
    setLoading(true)
    try {
      console.log('values', values)
      console.log('info', info)
      // 1. 创建组织
      const createRes = await createCompany({
        company_name: values.company_name,
        user_display_name: values.user_display_name,
        domain_name: location.origin,
        refresh_token: info.refreshToken,
        user_id: info.userInfo.id,
      })

      // 获取新创建的组织 ID
      const uinData = createRes?.uin
      const newUinId = uinData?.ID

      let finalUinList = info.uinList
      let finalRefreshToken = info.refreshToken
      let finalUserInfo = info.userInfo
      let finalJwtToken = ''

      // 2. 如果有用户名和密码，重新调用登录接口获取完整的 uinList
      if (info.username && info.password) {
        const encryptedPassword = encryptPassword(info.password)
        const loginBody = {
          username: info.username,
          password: encryptedPassword,
          domain_name: location.origin,
          // domain_name: 'https://example.com', //为了本地开发登录页面，临时设置的，开发完之后需要注释掉这行代码，恢复上一行的代码
        }
        const loginRes =
          version === 'custom'
            ? await loginByPasswordPrivate(loginBody)
            : await loginByPassword(loginBody)

        // 映射 uinList
        finalUinList = loginRes.uin.map((x: any) => {
          return {
            id: x.uin.ID,
            role: x.role,
            uinName: x.uin.Name,
            companyName: x.company_name,
            companyStatus: x.company_status, // 组织状态
            uinStatus: x.uin.UinStatus, // 身份状态
            subjectType: x.uin.SubjectType, // company 表示企业、individual 表示个人
            subjectId: x.uin.SubjectID, // 企业 id，个人为 0
            logo: x.company_logo, // 组织 logo
          }
        })

        finalRefreshToken = loginRes.refresh_token
        finalUserInfo = {
          id: loginRes.user_id,
          avatar: loginRes.user_info.avatar_url,
          name: loginRes.user_info.name,
          uinId: loginRes.user_info.uin,
          loginWay: loginRes.login_way,
        }
      } else {
        // 微信登录：将新创建的组织添加到 uinList 中
        // createCompany 接口返回的数据结构需要映射成和 uinList 一致的格式
        const uinData = createRes?.uin
        if (uinData) {
          const newUin = {
            id: uinData.ID,
            role: 'sys_admin', // 创建者默认为管理员
            uinName: uinData.Name || values.company_name,
            companyName: values.company_name,
            companyStatus: uinData.CompanyStatus || 'passed',
            uinStatus: uinData.UinStatus || 'normal',
            subjectType: uinData.SubjectType || 'company',
            subjectId: uinData.SubjectID || uinData.ID,
            logo: uinData.Logo || '',
          }
          // 检查是否已存在，避免重复添加
          const exists = finalUinList.some((u: any) => u.id === newUin.id)
          if (!exists) {
            finalUinList = [...finalUinList, newUin]
          }
        }
      }

      // 3. 选择新创建的组织身份完成登录
      const chooseBody = {
        login_way: finalUserInfo.loginWay,
        refresh_token: finalRefreshToken,
        uin_id: newUinId,
        user_id: finalUserInfo.id,
      }

      const chooseRes = await chooseUin(chooseBody)
      finalJwtToken = chooseRes.jwt_token

      // 4. 更新 localStorage
      localStore.setLogin({
        token: finalJwtToken,
        userInfo: {
          ...finalUserInfo,
          uinId: newUinId,
        },
        uinList: finalUinList,
      })

      messageApi.success(tM('loginSuccess'))
      onLogin()
    } catch (error: any) {
      console.log('error', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className='w-full h-60 flex flex-col pt-10'>
      <Form
        form={form}
        onFinish={handleCreateAndLogin}
        layout='vertical'
        className='w-full'
      >
        <Form.Item
          label={t('other.organizationName' as any)}
          name='company_name'
          rules={[
            { required: true, message: t('other.pleaseEnterOrganizationName') },
            { max: 50, message: t('other.organizationNameMaxLength') },
          ]}
        >
          <Input
            placeholder={t('other.pleaseEnterOrganizationName')}
            size='large'
            maxLength={50}
          />
        </Form.Item>

        <Form.Item
          label={t('other.userNickname' as any)}
          name='user_display_name'
          rules={[
            { required: true, message: t('other.pleaseEnterUserNickname') },
            { max: 20, message: t('other.userNicknameMaxLength') },
          ]}
          className='mt-8'
        >
          <Input
            placeholder={t('other.pleaseEnterUserNickname')}
            size='large'
            maxLength={20}
          />
        </Form.Item>

        <Form.Item className='mb-2'>
          <Button
            type='primary'
            htmlType='submit'
            size='large'
            loading={loading}
            className={cn(
              'w-full mt-6',
              isButtonDisabled && 'opacity-50 cursor-not-allowed',
            )}
          >
            {t('other.login')}
          </Button>
        </Form.Item>
      </Form>
      <Agreement className='mx-auto' />
    </div>
  )
}

export default CreateCompany
