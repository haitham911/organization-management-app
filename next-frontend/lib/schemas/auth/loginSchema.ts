import {z} from 'zod';

export const LoginSchema = z.object({
    email: z.string().email(),
})

export type LoginForm = z.infer<typeof LoginSchema>;