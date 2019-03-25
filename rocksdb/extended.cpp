/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

#include "extended.h"

#include "rocksdb/c.h"
#include "rocksdb/statistics.h"
#include "rocksdb/options.h"

using rocksdb::Statistics;
using rocksdb::HistogramData;
using rocksdb::StatsLevel;
using rocksdb::Options;
using std::shared_ptr;

extern "C" {

struct rocksdb_statistics_t { std::shared_ptr<Statistics> rep; };
struct rocksdb_histogram_data_t { rocksdb::HistogramData* rep; };
struct rocksdb_options_t { Options rep; };

rocksdb_statistics_t* rocksdb_create_statistics() {
    rocksdb_statistics_t* result = new rocksdb_statistics_t;
    result->rep = rocksdb::CreateDBStatistics();
    return result;
}

void rocksdb_options_set_statistics(
    rocksdb_options_t* opts, 
    rocksdb_statistics_t* stats) {
    if (stats) {
        opts->rep.statistics = stats->rep;
    }
}

rocksdb_stats_level_t rocksdb_statistics_stats_level(
    rocksdb_statistics_t* stats) {
        return static_cast<rocksdb_stats_level_t>(stats->rep->stats_level_);
}

void rocksdb_statistics_set_stats_level(
    rocksdb_statistics_t* stats,
    rocksdb_stats_level_t level) {
        stats->rep->stats_level_ = static_cast<StatsLevel>(level);
}

void rocksdb_statistics_reset(
    rocksdb_statistics_t* stats) {
        stats->rep->Reset();
}

uint64_t rocksdb_statistics_get_ticker_count(
    rocksdb_statistics_t* stats, 
    rocksdb_tickers_t ticker_type) {
        return stats->rep->getTickerCount(ticker_type);
}

uint64_t rocksdb_statistics_get_and_reset_ticker_count(
    rocksdb_statistics_t* stats, 
    rocksdb_tickers_t ticker_type) {
        return stats->rep->getAndResetTickerCount(ticker_type);
}

void rocksdb_statistics_destroy(rocksdb_statistics_t* stats) {
    delete stats;
}

void rocksdb_statistics_histogram_data(
    const rocksdb_statistics_t* stats, 
    rocksdb_histograms_t type, 
    const rocksdb_histogram_data_t* data) {
        stats->rep->histogramData(type, data->rep);
}

// Histogram

rocksdb_histogram_data_t* rocksdb_histogram_create_data() {
    rocksdb_histogram_data_t* result = new rocksdb_histogram_data_t;
    rocksdb::HistogramData hData;
    result->rep = &hData;
    return result;
}

double rocksdb_histogram_get_average(rocksdb_histogram_data_t* data) {
    return data->rep->average;
}

double rocksdb_histogram_get_median(rocksdb_histogram_data_t* data) {
    return data->rep->median;
}

double rocksdb_histogram_get_percentile95(rocksdb_histogram_data_t* data) {
    return data->rep->percentile95;
}

double rocksdb_histogram_get_percentile99(rocksdb_histogram_data_t* data) {
    return data->rep->percentile99;
}

double rocksdb_histogram_get_stdev(rocksdb_histogram_data_t* data) {
    return data->rep->standard_deviation;
}

double rocksdb_histogram_get_max(rocksdb_histogram_data_t* data) {
    return data->rep->max;
}

uint64_t rocksdb_histogram_get_count(rocksdb_histogram_data_t* data) {
    return data->rep->count;
}

uint64_t rocksdb_histogram_get_sum(rocksdb_histogram_data_t* data) {
    return data->rep->sum;
}

void rocksdb_histogram_data_destroy(rocksdb_histogram_data_t* data);

void rocksdb_histogram_data_destroy(rocksdb_histogram_data_t* data) {
    delete data;
}

} // end extern "C"
