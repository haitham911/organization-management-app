import { BASE_URL } from "@/config/config";
import { TNewOrganization, TOrganizationList } from "@/lib/types/organizationTypes";


export const getOrganizationsList = async (): Promise<TOrganizationList[]> => {
    const response =  await fetch(BASE_URL + '/organizations', {
        cache :'no-store'
    })
    const data = await response.json()
    return data
}
export const newOrganization = async (data: TNewOrganization) => {
    const response = await fetch(BASE_URL + '/organizations', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    })
    return response
}