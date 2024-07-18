import {z } from 'zod';

export const newOrganizationSchema = z.object({
  name: z.string().min(1).max(100),
  email: z.string().email(),
});

export type NewOrganizationForm = z.infer<typeof newOrganizationSchema>;