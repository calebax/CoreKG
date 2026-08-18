// 公司规模可选项
export const companyScale = Array.from(
  { length: 6 },
  (_, i) => `version.companyScaleOpt.option${i + 1}`,
)

// 行业信息可选项
const industryConfig = [
  { key: 'technologyInternet', optionsCount: 4 },
  { key: 'finance', optionsCount: 4 },
  { key: 'medicalHealth', optionsCount: 4 },
  { key: 'education', optionsCount: 4 },
  { key: 'retailConsumerGoods', optionsCount: 4 },
  { key: 'manufacturing', optionsCount: 4 },
  { key: 'professionalServices', optionsCount: 5 },
  { key: 'realEstateConstruction', optionsCount: 3 },
  { key: 'mediaEntertainment', optionsCount: 3 },
  { key: 'logisticsTransport', optionsCount: 3 },
  { key: 'energyUtilities', optionsCount: 3 },
  { key: 'governmentNonProfit', optionsCount: 2 },
  { key: 'other', optionsCount: 2 },
]

export const companyIndustry = industryConfig.map(({ key, optionsCount }) => ({
  label: `version.companyIndustry.${key}.label`,
  options: Array.from(
    { length: optionsCount },
    (_, i) => `version.companyIndustry.${key}.option${i + 1}`,
  ),
}))
