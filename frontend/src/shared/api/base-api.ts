import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

export const baseApi = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: '/api/v1/',
    timeout: 10_000,
    prepareHeaders: headers => { headers.set('Accept', 'application/json'); return headers; },
  }),
  refetchOnReconnect: true,
  keepUnusedDataFor: 60,
  endpoints: () => ({}),
});
