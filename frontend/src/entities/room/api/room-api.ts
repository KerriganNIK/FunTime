import { baseApi } from '@/shared/api';
import type { RoomSnapshot } from '../model/types';

export const roomApi = baseApi.enhanceEndpoints({ addTagTypes: ['Room'] }).injectEndpoints({
  endpoints: build => ({
    room: build.query<RoomSnapshot, { code: string; screen?: boolean }>({
      query: ({ code, screen }) => `rooms/${encodeURIComponent(code)}${screen ? '/screen' : ''}`,
      providesTags: (_data, _error, { code }) => [{ type: 'Room', id: code }],
      keepUnusedDataFor: 0,
    }),
    createRoom: build.mutation<{ code: string }, { name: string; gameId: string }>({ query: body => ({ url: 'rooms', method: 'POST', body }) }),
    joinRoom: build.mutation<{ code: string }, { code: string; name: string }>({
      query: ({ code, name }) => ({ url: `rooms/${encodeURIComponent(code)}/join`, method: 'POST', body: { name } }),
      invalidatesTags: (_data, _error, { code }) => [{ type: 'Room', id: code }],
    }),
    roomCommand: build.mutation<RoomSnapshot, { code: string; kind: string; payload: unknown; requestId: string }>({
      query: ({ code, ...body }) => ({ url: `rooms/${encodeURIComponent(code)}/commands`, method: 'POST', body }),
      async onQueryStarted({ code }, { dispatch, queryFulfilled }) {
        try { const { data } = await queryFulfilled; dispatch(roomApi.util.upsertQueryData('room', { code, screen: false }, data)); } catch { /* Mutation error is shown by the caller. */ }
      },
      invalidatesTags: (_data, _error, { code }) => [{ type: 'Room', id: code }],
    }),
  }),
});
export const { useRoomQuery, useCreateRoomMutation, useJoinRoomMutation, useRoomCommandMutation } = roomApi;
