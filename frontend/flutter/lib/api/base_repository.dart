import 'package:dio/dio.dart';

import 'package:online_cource_app/api/api_client.dart';
import 'package:online_cource_app/api/models/models.dart';
import 'package:online_cource_app/api/repositories/auth_repository.dart';

/// Shared HTTP helpers and error mapping for REST repositories.
abstract class BaseRepository {
  final ApiClient client;

  BaseRepository(this.client);

  Future<Map<String, dynamic>> getJson(
    String path, {
    Map<String, dynamic>? query,
  }) async {
    try {
      final res = await client.dio.get<Map<String, dynamic>>(
        path,
        queryParameters: query,
      );
      return res.data!;
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<List<dynamic>> getList(String path, {Map<String, dynamic>? query}) async {
    try {
      final res = await client.dio.get<List<dynamic>>(
        path,
        queryParameters: query,
      );
      return res.data ?? [];
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<Map<String, dynamic>> postJson(
    String path,
    Map<String, dynamic> body,
  ) async {
    try {
      final res = await client.dio.post<Map<String, dynamic>>(path, data: body);
      return res.data!;
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<Map<String, dynamic>> patchJson(
    String path,
    Map<String, dynamic> body,
  ) async {
    try {
      final res = await client.dio.patch<Map<String, dynamic>>(path, data: body);
      return res.data!;
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  Future<void> delete(String path) async {
    try {
      await client.dio.delete<void>(path);
    } on DioException catch (e) {
      throw toApiException(e);
    }
  }

  ApiException toApiException(DioException e) {
    final data = e.response?.data;
    final apiError = data is Map<String, dynamic>
        ? ApiError.fromJson(data)
        : ApiError(
            error: 'network',
            message: e.message ?? 'Network error',
          );
    return ApiException(e.response?.statusCode, apiError);
  }
}
