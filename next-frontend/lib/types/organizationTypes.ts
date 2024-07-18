export type TOrganizationList = TOrganization[]

export type TOrganization  = {
  ID: number
  CreatedAt: string
  UpdatedAt: string
  DeletedAt: any
  name: string
  email: string
  stripe_customer_id: string
  Users: any[]
  subscriptions: any
}

export type TNewOrganization = {
  email: string
  name: string
}

