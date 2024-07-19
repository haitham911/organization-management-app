export const getProducts = async () => {
    console.log('process.env.NEXT_PUBLIC_API_URL', process.env.NEXT_PUBLIC_API_URL);
    const baseUrl = process.env.NEXT_PUBLIC_API_URL + '/products';
    return await fetch(baseUrl)
}