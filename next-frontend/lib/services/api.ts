
import axios, {AxiosRequestConfig} from 'axios';


const config: AxiosRequestConfig = {
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  maxRedirects: 0,
};

const api = axios.create(config);

export default api;
